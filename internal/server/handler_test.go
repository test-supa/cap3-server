package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/your-org/chameleon-c2/internal/types"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "chameleon-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	tmpFile.Close()

	cfg := &Config{
		ListenAddr:   ":0",
		DBPath:       tmpFile.Name(),
		MasterSecret: "test-master-secret-for-testing",
		OfflineAfter: 120,
	}

	srv, err := New(cfg)
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to create server: %v", err)
	}

	t.Cleanup(func() {
		srv.Stop()
		os.Remove(tmpFile.Name())
	})

	return srv
}

func TestHealthEndpoint(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}

func TestListDevicesEndpoint(t *testing.T) {
	srv := setupTestServer(t)

	// Initially no devices
	req := httptest.NewRequest("GET", "/api/devices", nil)
	w := httptest.NewRecorder()
	srv.handleListDevices(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var devices []types.Device
	json.NewDecoder(w.Body).Decode(&devices)
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestStatsEndpoint(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	srv.handleStats(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var stats map[string]interface{}
	json.NewDecoder(w.Body).Decode(&stats)

	if stats["total_devices"] != float64(0) {
		t.Errorf("expected 0 total devices, got %v", stats["total_devices"])
	}
}

func TestDeviceDetailNotFound(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/devices/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.handleDeviceDetail(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSendCommandMissingFields(t *testing.T) {
	srv := setupTestServer(t)

	tests := []struct {
		name string
		body string
		code int
	}{
		{"empty body", `{}`, 400},
		{"missing device_id", `{"command":"test"}`, 400},
		{"missing command", `{"device_id":"dev-1"}`, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/command", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			srv.handleSendCommand(w, req)

			if w.Code != tt.code {
				t.Errorf("expected %d, got %d", tt.code, w.Code)
			}
		})
	}
}

func TestSendCommandQueued(t *testing.T) {
	srv := setupTestServer(t)

	body := `{"device_id":"dev-1","command":"start_sweep","params":{"duration":300,"targets":["all"]}}`
	req := httptest.NewRequest("POST", "/api/command", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleSendCommand(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"] != "queued" {
		t.Errorf("expected queued, got %s", resp["status"])
	}
	if resp["command_id"] == "" {
		t.Error("expected non-empty command_id")
	}
}

func TestSendCommandMethodNotAllowed(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/command", nil)
	w := httptest.NewRecorder()
	srv.handleSendCommand(w, req)

	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestDashboardEndpoint(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.handleDashboard(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected HTML content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestDashboard404(t *testing.T) {
	srv := setupTestServer(t)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.handleDashboard(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestWebSocketUpgrade(t *testing.T) {
	srv := setupTestServer(t)

	server := httptest.NewServer(http.HandlerFunc(srv.handleWS))
	defer server.Close()

	// Just verify the endpoint exists by checking the response
	// (WebSocket upgrade will fail with HTTP request, but should return 400, not 404)
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// WebSocket upgrade should fail with 400 since we sent HTTP, not WS
	if resp.StatusCode != 400 {
		t.Logf("expected 400 for non-WebSocket request, got %d", resp.StatusCode)
	}
}
