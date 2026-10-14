package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestPersistentAIAgentClientStoreReloadsDeviceDaemonList(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev"}
	writer, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	reader, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("open stale reader: %v", err)
	}
	reader.snapshotReloadInterval = 0

	enrollment, err := writer.EnrollDeviceCredential(ctx, principal, "workspace-dev",
		EnrollDeviceRequest{DisplayName: "Development Mac"})
	if err != nil {
		t.Fatalf("EnrollDeviceCredential: %v", err)
	}
	startedAt := time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC)
	if _, err := writer.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          "daemon-device-list-staging",
		DeviceID:          enrollment.DeviceID,
		DeviceDisplayName: "Development Mac",
		Profile:           "staging",
		PID:               8822,
		UptimeSeconds:     600,
		StartedAt:         startedAt,
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID: "daemon-device-list-staging:codex",
			Kind:      RuntimeKindCodex,
		}},
	}); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot: %v", err)
	}

	list, err := reader.ListAIAgentDeviceDaemons(ctx, principal, enrollment.DeviceID)
	if err != nil {
		t.Fatalf("ListAIAgentDeviceDaemons: %v", err)
	}
	if len(list.Daemons) != 1 {
		t.Fatalf("daemon list after reload = %+v", list.Daemons)
	}
	daemon := list.Daemons[0]
	if daemon.DaemonID != "daemon-device-list-staging" || daemon.PID != 8822 ||
		daemon.Profile != "staging" || !daemon.StartedAt.Equal(startedAt) {
		t.Fatalf("reloaded daemon facts = %+v", daemon)
	}
}
