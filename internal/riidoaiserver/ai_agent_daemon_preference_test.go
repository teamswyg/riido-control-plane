package riidoaiserver

import (
	"testing"
	"time"
)

func TestDaemonBetterForDevicePrefersFreshestLastSeenThenStartedAt(t *testing.T) {
	base := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	current := DeviceDaemonRecord{
		DeviceID:   "device-1",
		LastSeenAt: base,
		StartedAt:  base.Add(-time.Hour),
	}
	if !daemonBetterForDevice(current, DeviceDaemonRecord{}) {
		t.Fatal("first daemon should be preferred over empty current")
	}
	if !daemonBetterForDevice(DeviceDaemonRecord{
		DeviceID:   "device-1",
		LastSeenAt: base.Add(time.Second),
		StartedAt:  base.Add(-2 * time.Hour),
	}, current) {
		t.Fatal("newer last_seen daemon should be preferred")
	}
	if !daemonBetterForDevice(DeviceDaemonRecord{
		DeviceID:   "device-1",
		LastSeenAt: base,
		StartedAt:  base,
	}, current) {
		t.Fatal("same last_seen should prefer newer started_at")
	}
	if daemonBetterForDevice(DeviceDaemonRecord{
		DeviceID:   "device-1",
		LastSeenAt: base.Add(-time.Second),
		StartedAt:  base,
	}, current) {
		t.Fatal("older last_seen daemon should not be preferred")
	}
}
