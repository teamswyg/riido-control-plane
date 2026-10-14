package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestAIAgentDaemonRuntimeSnapshotIgnoresOlderDaemonInstance(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	startedAt := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	principal := AuthorizationResult{PrincipalID: "user-daemon-stale", WorkspaceID: "workspace-daemon-stale"}

	current, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          "daemon-stale",
		DeviceID:          "device-stale",
		DeviceDisplayName: "Stale Guard Mac",
		PID:               200,
		StartedAt:         startedAt,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:      "device-stale:codex",
			Kind:           RuntimeKindCodex,
			Availability:   RuntimeAvailabilityOnline,
			DetectionState: RuntimeDetectionStateDetected,
		}},
	})
	if err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot current: %v", err)
	}
	requireDaemonPID(t, current.Daemon, 200)

	stale, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          "daemon-stale",
		DeviceID:          "device-stale",
		DeviceDisplayName: "Stale Guard Mac",
		PID:               100,
		StartedAt:         startedAt.Add(-time.Minute),
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:      "device-stale:claude",
			Kind:           RuntimeKindClaudeCode,
			Availability:   RuntimeAvailabilityOffline,
			DetectionState: RuntimeDetectionStateMissing,
		}},
	})
	if err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot stale: %v", err)
	}
	requireDaemonPID(t, stale.Daemon, 200)

	devices, err := store.ListAIAgentDevices(ctx, principal)
	if err != nil {
		t.Fatalf("ListAIAgentDevices: %v", err)
	}
	device := requireDeviceByID(t, devices.Devices, "device-stale")
	if _, ok := runtimeByID(device.Runtimes, "device-stale:codex"); !ok {
		t.Fatalf("current runtime missing after stale snapshot: %+v", device.Runtimes)
	}
	if _, ok := runtimeByID(device.Runtimes, "device-stale:claude"); ok {
		t.Fatalf("stale runtime overwrote current snapshot: %+v", device.Runtimes)
	}
}
