package riidoaiserver

import "testing"

func requireDeviceByID(t *testing.T, devices []DeviceRecord, deviceID string) DeviceRecord {
	t.Helper()
	for _, device := range devices {
		if device.DeviceID == deviceID {
			return device
		}
	}
	t.Fatalf("device %s not found: %+v", deviceID, devices)
	return DeviceRecord{}
}

func requireDaemonPID(t *testing.T, daemon DeviceDaemonRecord, pid int) {
	t.Helper()
	if daemon.PID != pid {
		t.Fatalf("daemon pid=%d, want %d", daemon.PID, pid)
	}
}
