package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestAIAgentDaemonRuntimeSnapshotRejectsUnexpectedProfile(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{
		PrincipalID: "user-profiled",
		WorkspaceID: "workspace-profiled",
	}
	deviceID := "device-profiled-reject"

	syncProfileRuntime(t, store, ctx, principal,
		profileRuntimeSnapshot(deviceID, "staging", RuntimeKindCodex))
	syncProfileRuntime(t, store, ctx, principal,
		profileRuntimeSnapshot(deviceID, "production", RuntimeKindClaudeCode))
	if err := store.ConfigureDaemonProfile("production"); err != nil {
		t.Fatalf("ConfigureDaemonProfile: %v", err)
	}

	_, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal,
		profileRuntimeSnapshot(deviceID, "staging", RuntimeKindCursor))
	if !errors.Is(err, ErrDaemonProfileMismatch) {
		t.Fatalf("SyncAIAgentDaemonRuntimeSnapshot err=%v, want mismatch", err)
	}
	assertOnlyProfileRuntime(t, store, deviceID, "production")
	if got := countDaemonsForDevice(store, deviceID); got != 1 {
		t.Fatalf("daemon records for %s = %d, want 1", deviceID, got)
	}
}
