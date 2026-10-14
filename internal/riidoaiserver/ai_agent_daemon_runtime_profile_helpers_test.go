package riidoaiserver

import (
	"context"
	"testing"
)

func syncProfileRuntime(t *testing.T, store *DevelopmentAIAgentClientStore, ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) {
	t.Helper()
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot %s: %v", req.Profile, err)
	}
}

func requireAgentDaemonBinding(t *testing.T, store *DevelopmentAIAgentClientStore, agentID, daemonID string) {
	t.Helper()
	binding, ok := store.LookupAgent(agentID)
	if !ok {
		t.Fatalf("agent %s binding missing", agentID)
	}
	if binding.DaemonID != daemonID {
		t.Fatalf("agent %s daemon_id = %s, want %s", agentID, binding.DaemonID, daemonID)
	}
}

func countDaemonsForDevice(store *DevelopmentAIAgentClientStore, deviceID string) int {
	store.mu.Lock()
	defer store.mu.Unlock()
	count := 0
	for _, daemon := range store.daemons {
		if daemon.DeviceID == deviceID {
			count++
		}
	}
	return count
}
