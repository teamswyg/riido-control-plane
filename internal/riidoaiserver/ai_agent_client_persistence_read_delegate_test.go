package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentClientStoreReadDelegatesDoNotSave(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	baselineSaves := snapshots.saves

	fixtures, err := store.ListAIAgentOnboardingFixtures(ctx, principal)
	if err != nil || len(fixtures.Fixtures) == 0 {
		t.Fatalf("ListAIAgentOnboardingFixtures = %+v, %v", fixtures, err)
	}
	assignable, err := store.ListAIAgentTaskAssignableAgents(ctx, principal, "task-read-delegate")
	if err != nil || len(assignable.Agents) == 0 {
		t.Fatalf("ListAIAgentTaskAssignableAgents = %+v, %v", assignable, err)
	}
	profiles, err := store.ListWorkspaceAssignedAgentProfiles(ctx, principal)
	if err != nil || len(profiles.AssignedAgentProfiles) == 0 {
		t.Fatalf("ListWorkspaceAssignedAgentProfiles = %+v, %v", profiles, err)
	}
	if _, err := store.GetAIAgentEditability(ctx, principal, "agent-owned-codex"); err != nil {
		t.Fatalf("GetAIAgentEditability: %v", err)
	}
	if snapshots.saves != baselineSaves {
		t.Fatalf("read delegates changed snapshot saves: %d -> %d", baselineSaves, snapshots.saves)
	}
}

func TestPersistentAIAgentClientStoreStreamDelegatesDoNotSave(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	baselineSaves := snapshots.saves

	if _, err := store.GetAIAgentTaskThreadStreamSubscription(ctx, principal, "task-read-stream"); err != nil {
		t.Fatalf("GetAIAgentTaskThreadStreamSubscription: %v", err)
	}
	if _, err := store.AIAgentClientEvents(ctx, principal); err != nil {
		t.Fatalf("AIAgentClientEvents: %v", err)
	}
	_, _, unsubscribe, err := store.SubscribeAIAgentClientEvents(ctx, principal)
	if err != nil {
		t.Fatalf("SubscribeAIAgentClientEvents: %v", err)
	}
	unsubscribe()
	if snapshots.saves != baselineSaves {
		t.Fatalf("stream delegates changed snapshot saves: %d -> %d", baselineSaves, snapshots.saves)
	}
}

func openPersistentAIAgentClientStoreForDelegateTest(
	t *testing.T,
	snapshots *memoryAIAgentClientSnapshotStore,
) *PersistentAIAgentClientStore {
	t.Helper()
	store, err := OpenPersistentAIAgentClientStore(
		context.Background(),
		NewDevelopmentAIAgentClientStore(),
		snapshots,
	)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	return store
}
