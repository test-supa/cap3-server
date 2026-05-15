package db

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/chameleon-c2/internal/types"
)

func setupTestDB(t *testing.T) *Store {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "chameleon-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	tmpFile.Close()

	store, err := New(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to create store: %v", err)
	}

	t.Cleanup(func() {
		store.Close()
		os.Remove(tmpFile.Name())
	})

	return store
}

func TestNewStore(t *testing.T) {
	store := setupTestDB(t)
	if store == nil {
		t.Fatal("store should not be nil")
	}
}

func TestRegisterAndGetDevice(t *testing.T) {
	store := setupTestDB(t)

	info := types.DeviceInfo{
		DeviceID:       "test-device-1",
		DeviceName:     "Samsung Galaxy S24",
		Manufacturer:   "Samsung",
		Model:          "SM-S928B",
		AndroidVersion: "14",
		APiLevel:       34,
		IPAddress:      "192.168.1.100",
	}

	err := store.RegisterDevice(info)
	if err != nil {
		t.Fatalf("register device failed: %v", err)
	}

	device, err := store.GetDevice("test-device-1")
	if err != nil {
		t.Fatalf("get device failed: %v", err)
	}

	if device.DeviceID != "test-device-1" {
		t.Errorf("expected device_id test-device-1, got %s", device.DeviceID)
	}
	if device.Model != "SM-S928B" {
		t.Errorf("expected model SM-S928B, got %s", device.Model)
	}
	if device.Status != "online" {
		t.Errorf("expected status online, got %s", device.Status)
	}
}

func TestRegisterDeviceTwice(t *testing.T) {
	store := setupTestDB(t)

	info := types.DeviceInfo{
		DeviceID:       "test-device",
		DeviceName:     "Original",
		Manufacturer:   "Google",
		Model:          "Pixel 8",
		AndroidVersion: "14",
		APiLevel:       34,
	}

	err := store.RegisterDevice(info)
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	// Second registration with different info should update
	info.DeviceName = "Updated"
	info.AndroidVersion = "15"
	info.APiLevel = 35

	err = store.RegisterDevice(info)
	if err != nil {
		t.Fatalf("second register failed: %v", err)
	}

	device, err := store.GetDevice("test-device")
	if err != nil {
		t.Fatalf("get device failed: %v", err)
	}

	if device.DeviceName != "Updated" {
		t.Errorf("expected Updated, got %s", device.DeviceName)
	}
	if device.APiLevel != 35 {
		t.Errorf("expected API 35, got %d", device.APiLevel)
	}
}

func TestListDevices(t *testing.T) {
	store := setupTestDB(t)

	devices, err := store.ListDevices()
	if err != nil {
		t.Fatalf("list devices failed: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}

	store.RegisterDevice(types.DeviceInfo{DeviceID: "dev-1", Manufacturer: "Samsung"})
	store.RegisterDevice(types.DeviceInfo{DeviceID: "dev-2", Manufacturer: "Google"})

	devices, err = store.ListDevices()
	if err != nil {
		t.Fatalf("list devices failed: %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestUpdateHeartbeat(t *testing.T) {
	store := setupTestDB(t)

	store.RegisterDevice(types.DeviceInfo{DeviceID: "dev-heartbeat"})

	// Wait a moment, then heartbeat
	time.Sleep(10 * time.Millisecond)
	err := store.UpdateHeartbeat("dev-heartbeat")
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
}

func TestSetDeviceOffline(t *testing.T) {
	store := setupTestDB(t)

	store.RegisterDevice(types.DeviceInfo{DeviceID: "dev-offline"})
	err := store.SetDeviceOffline("dev-offline")
	if err != nil {
		t.Fatalf("set offline failed: %v", err)
	}

	device, _ := store.GetDevice("dev-offline")
	if device.Status != "offline" {
		t.Errorf("expected offline, got %s", device.Status)
	}
}

func TestCommandQueue(t *testing.T) {
	store := setupTestDB(t)

	cmd := types.CommandMessage{
		CommandID: "cmd-001",
		Command:   "start_sweep",
	}

	err := store.EnqueueCommand("dev-1", cmd)
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	cmds, err := store.GetPendingCommands("dev-1")
	if err != nil {
		t.Fatalf("get pending failed: %v", err)
	}
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].CommandID != "cmd-001" {
		t.Errorf("expected cmd-001, got %s", cmds[0].CommandID)
	}
}

func TestCommandQueueOrder(t *testing.T) {
	store := setupTestDB(t)

	store.EnqueueCommand("dev-1", types.CommandMessage{CommandID: "cmd-001", Command: "first"})
	store.EnqueueCommand("dev-1", types.CommandMessage{CommandID: "cmd-002", Command: "second"})

	cmds, _ := store.GetPendingCommands("dev-1")
	if len(cmds) < 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if cmds[0].CommandID != "cmd-001" {
		t.Errorf("expected cmd-001 first, got %s", cmds[0].CommandID)
	}
}

func TestUpdateCommandStatus(t *testing.T) {
	store := setupTestDB(t)

	store.EnqueueCommand("dev-1", types.CommandMessage{CommandID: "cmd-001", Command: "test"})

	err := store.UpdateCommandStatus("cmd-001", "done", "success")
	if err != nil {
		t.Fatalf("update status failed: %v", err)
	}

	// After update, should no longer be pending
	cmds, _ := store.GetPendingCommands("dev-1")
	if len(cmds) != 0 {
		t.Errorf("expected 0 pending commands, got %d", len(cmds))
	}
}

func TestLogDataReceived(t *testing.T) {
	store := setupTestDB(t)

	err := store.LogDataReceived("dev-1", "credential", "test credential data")
	if err != nil {
		t.Fatalf("log data failed: %v", err)
	}
}

func TestMarkOfflineDevices(t *testing.T) {
	store := setupTestDB(t)

	store.RegisterDevice(types.DeviceInfo{DeviceID: "dev-old"})
	store.RegisterDevice(types.DeviceInfo{DeviceID: "dev-new"})

	// Mark devices with old last_seen as offline
	err := store.MarkOfflineDevices(0 * time.Second)
	if err != nil {
		t.Fatalf("mark offline failed: %v", err)
	}
}

func TestGetDeviceNotFound(t *testing.T) {
	store := setupTestDB(t)

	_, err := store.GetDevice("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
}

func TestCloseStore(t *testing.T) {
	store := setupTestDB(t)
	err := store.Close()
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
}
