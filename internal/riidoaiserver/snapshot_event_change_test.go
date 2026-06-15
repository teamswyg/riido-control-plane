package riidoaiserver

import (
	"testing"
	"time"
)

func TestDeviceRuntimeSnapshotChangedForEvent(t *testing.T) {
	t0 := time.Now().UTC()
	base := DeviceRecord{
		DeviceID: "dev_x", OwnerPrincipalID: "u", DisplayName: "Mac",
		DaemonLastSeenAt: t0,
		Runtimes: []RuntimeRecord{
			{RuntimeID: "dev_x:claude", Kind: RuntimeKindClaudeCode, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected, LastDetectedAt: t0},
		},
	}
	// No prior → changed (first snapshot must publish).
	if !deviceRuntimeSnapshotChangedForEvent(DeviceRecord{}, false, base) {
		t.Fatal("first snapshot should be treated as changed")
	}
	// Heartbeat: only liveness timestamps advanced → NOT changed (suppressed).
	hb := base
	hb.DaemonLastSeenAt = t0.Add(5 * time.Second)
	hb.Runtimes = []RuntimeRecord{{RuntimeID: "dev_x:claude", Kind: RuntimeKindClaudeCode, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected, LastDetectedAt: t0.Add(5 * time.Second)}}
	if deviceRuntimeSnapshotChangedForEvent(base, true, hb) {
		t.Fatal("pure heartbeat (only timestamps) should NOT be a change")
	}
	// Runtime went offline → changed.
	off := base
	off.Runtimes = []RuntimeRecord{{RuntimeID: "dev_x:claude", Kind: RuntimeKindClaudeCode, Availability: RuntimeAvailabilityOffline, DetectionState: RuntimeDetectionStateMissing, LastDetectedAt: t0}}
	if !deviceRuntimeSnapshotChangedForEvent(base, true, off) {
		t.Fatal("availability change should be a change")
	}
	version := base
	version.Runtimes = []RuntimeRecord{{RuntimeID: "dev_x:claude", Kind: RuntimeKindClaudeCode, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected, ProviderVersion: "2.1.142 (Claude Code)", LastDetectedAt: t0}}
	if !deviceRuntimeSnapshotChangedForEvent(base, true, version) {
		t.Fatal("provider_version change should be a change")
	}
}

func TestDaemonStatusChangedForEvent(t *testing.T) {
	t0 := time.Now().UTC()
	base := DeviceDaemonRecord{DeviceID: "dev_x", DaemonID: "dev_x", Profile: "local", PID: 100, Availability: DaemonAvailabilityOnline, ControlState: DaemonControlStateIdle, LastSeenAt: t0, UptimeSeconds: 10}
	if !daemonStatusChangedForEvent(DeviceDaemonRecord{}, false, base) {
		t.Fatal("first daemon status should be treated as changed")
	}
	hb := base
	hb.LastSeenAt = t0.Add(5 * time.Second)
	hb.UptimeSeconds = 15
	if daemonStatusChangedForEvent(base, true, hb) {
		t.Fatal("pure heartbeat (last_seen/uptime) should NOT be a change")
	}
	restarted := base
	restarted.PID = 200
	if !daemonStatusChangedForEvent(base, true, restarted) {
		t.Fatal("PID change (restart) should be a change")
	}
}
