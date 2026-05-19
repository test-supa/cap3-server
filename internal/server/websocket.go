package server

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/your-org/chameleon-c2/internal/types"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10 * 1024 * 1024 // 10MB
)

type Client struct {
	DeviceID   string
	Conn       *websocket.Conn
	Server     *Server
	mu         sync.Mutex
	done       chan struct{}
	IsAdmin    bool
	StreamFor  string // device_id this admin is streaming
}

func NewClient(conn *websocket.Conn, srv *Server) *Client {
	c := &Client{
		Conn:   conn,
		Server: srv,
		done:   make(chan struct{}),
	}
	conn.SetReadLimit(maxMessageSize)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	return c
}

func NewAdminClient(conn *websocket.Conn, srv *Server) *Client {
	c := NewClient(conn, srv)
	c.IsAdmin = true
	return c
}

func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}

func (c *Client) SendMessage(msg types.WSMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) SendBinary(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Client) SendCommand(cmd types.CommandMessage) error {
	return c.SendMessage(types.WSMessage{
		Type:      types.MsgCommand,
		CommandID: cmd.CommandID,
		Command:   cmd.Command,
		Params:    cmd.Params,
	})
}

func (c *Client) Close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsAdmin && c.StreamFor != "" {
		c.Server.adminHub.UnregisterViewer(c.StreamFor, c)
	}
	if c.DeviceID != "" {
		c.Server.hub.Unregister(c.DeviceID)
		c.Server.db.SetDeviceOffline(c.DeviceID)
	}
	c.Conn.Close()
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer c.Close()

	for {
		select {
		case <-c.done:
			return
		default:
			c.Conn.SetReadDeadline(time.Now().Add(pongWait))
			msgType, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					log.Printf("ws read error from %s: %v", c.DeviceID, err)
				}
				return
			}

			if msgType == websocket.BinaryMessage {
				// Binary frame from device — likely a stream frame
				c.Server.handleBinaryFrame(c, message)
				continue
			}

			var msg types.WSMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("invalid message from %s: %v", c.DeviceID, err)
				continue
			}

			c.Server.handleMessage(c, msg)
		}
	}
}

// ─── DEVICE HUB ────────────────────────────────────────────────────
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // deviceID -> client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) Register(deviceID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[deviceID] = client
}

func (h *Hub) Unregister(deviceID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, deviceID)
}

func (h *Hub) GetClient(deviceID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[deviceID]
}

func (h *Hub) SendToDevice(deviceID string, msg types.WSMessage) error {
	client := h.GetClient(deviceID)
	if client == nil {
		return fmt.Errorf("device %s not connected", deviceID)
	}
	return client.SendMessage(msg)
}

func (h *Hub) SendCommand(deviceID string, cmd types.CommandMessage) error {
	client := h.GetClient(deviceID)
	if client == nil {
		return fmt.Errorf("device %s not connected", deviceID)
	}
	return client.SendCommand(cmd)
}

func (h *Hub) BroadcastMessage(msg types.WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		client.SendMessage(msg)
	}
}

func (h *Hub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) ConnectedDeviceIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

// ─── ADMIN HUB (Browser viewers) ────────────────────────────────────
type AdminHub struct {
	mu       sync.RWMutex
	viewers  map[string][]*Client // deviceID -> admin clients viewing it
}

func NewAdminHub() *AdminHub {
	return &AdminHub{
		viewers: make(map[string][]*Client),
	}
}

func (ah *AdminHub) RegisterViewer(deviceID string, client *Client) {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	client.StreamFor = deviceID
	ah.viewers[deviceID] = append(ah.viewers[deviceID], client)
}

func (ah *AdminHub) UnregisterViewer(deviceID string, client *Client) {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	viewers := ah.viewers[deviceID]
	for i, v := range viewers {
		if v == client {
			ah.viewers[deviceID] = append(viewers[:i], viewers[i+1:]...)
			break
		}
	}
	if len(ah.viewers[deviceID]) == 0 {
		delete(ah.viewers, deviceID)
	}
}

func (ah *AdminHub) BroadcastFrame(deviceID string, frameData []byte) {
	ah.mu.RLock()
	defer ah.mu.RUnlock()
	for _, viewer := range ah.viewers[deviceID] {
		viewer.SendBinary(frameData)
	}
}

func (ah *AdminHub) ViewerCount(deviceID string) int {
	ah.mu.RLock()
	defer ah.mu.RUnlock()
	return len(ah.viewers[deviceID])
}
