package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentClientStoreDaemonDelegates(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	baselineSaves := snapshots.saves

	detail, err := store.GetAIAgentDaemon(ctx, principal, "agent-owned-codex")
	if err != nil || detail.Daemon.DeviceID != "device-dev-macbook" {
		t.Fatalf("GetAIAgentDaemon = %+v, %v", detail, err)
	}
	bindings, err := store.ListAIAgentDaemonAgentBindings(ctx, principal, detail.Daemon.DeviceID)
	if err != nil || len(bindings.Bindings) == 0 {
		t.Fatalf("ListAIAgentDaemonAgentBindings = %+v, %v", bindings, err)
	}
	changed, err := store.ReconcileAIAgentActiveThreadProjections(ctx, principal, "task-no-reader", nil)
	if err != nil || changed {
		t.Fatalf("ReconcileAIAgentActiveThreadProjections changed=%v err=%v", changed, err)
	}
	if snapshots.saves != baselineSaves {
		t.Fatalf("read daemon delegates changed saves: %d -> %d", baselineSaves, snapshots.saves)
	}
	assertPersistentDaemonControlsSave(t, ctx, store, principal, snapshots)
}

func assertPersistentDaemonControlsSave(
	t *testing.T,
	ctx context.Context,
	store *PersistentAIAgentClientStore,
	principal AuthorizationResult,
	snapshots *memoryAIAgentClientSnapshotStore,
) {
	t.Helper()
	before := snapshots.saves
	agentCmd, err := store.ControlAIAgentDaemon(
		ctx,
		principal,
		"agent-owned-codex",
		DaemonControlActionRestart,
		ControlDeviceDaemonRequest{Reason: "delegate test"},
	)
	if err != nil || agentCmd.Action != DaemonControlActionRestart {
		t.Fatalf("ControlAIAgentDaemon = %+v, %v", agentCmd, err)
	}
	if snapshots.saves <= before {
		t.Fatalf("agent daemon control did not save: %d <= %d", snapshots.saves, before)
	}
	before = snapshots.saves
	deviceCmd, err := store.ControlAIAgentDeviceDaemon(
		ctx,
		principal,
		"device-dev-macbook",
		DaemonControlActionStop,
		ControlDeviceDaemonRequest{Reason: "delegate test"},
	)
	if err != nil || deviceCmd.Action != DaemonControlActionStop {
		t.Fatalf("ControlAIAgentDeviceDaemon = %+v, %v", deviceCmd, err)
	}
	if snapshots.saves <= before {
		t.Fatalf("device daemon control did not save: %d <= %d", snapshots.saves, before)
	}
}
