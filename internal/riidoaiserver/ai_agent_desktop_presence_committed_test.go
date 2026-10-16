package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDesktopPresenceDoesNotPublishUncommittedMemory(t *testing.T) {
	ctx := context.Background()
	snapshots := grantsReadingSnapshotStore{&memoryAIAgentClientSnapshotStore{}}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	principal := profileColorTestPrincipal()
	store.mu.Lock()
	store.upsertDeviceConnectionGrantLocked("uncommitted-device", principal, time.Now())
	store.mu.Unlock()
	seen, err := store.DesktopPresenceLastSeenAt(ctx, principal)
	if err != nil || !seen.IsZero() {
		t.Fatalf("uncommitted memory published: seen=%t error=%v", !seen.IsZero(), err)
	}
}

func TestDesktopPresenceSurvivesStaleOrdinarySnapshotWriter(t *testing.T) {
	ctx := context.Background()
	fixture := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, fixture)
	stale, _ := newMembershipDynamoReplica(t, fixture)
	principal := profileColorTestPrincipal()
	if _, err := writer.ConnectAIAgentDevice(ctx, principal, "presence-device"); err != nil {
		t.Fatal(err)
	}
	before, err := writer.DesktopPresenceLastSeenAt(ctx, principal)
	if err != nil || before.IsZero() {
		t.Fatalf("committed connection missing: %v", err)
	}
	for range 2 {
		if _, err := stale.UpdateAIAgentConfiguration(ctx, principal, "agent-owned-claude", UpdateAgentConfigurationRequest{Name: "Unrelated update"}); err != nil {
			t.Fatal(err)
		}
	}
	fresh, _ := newMembershipDynamoReplica(t, fixture)
	after, err := fresh.DesktopPresenceLastSeenAt(ctx, principal)
	if err != nil || !after.Equal(before) {
		t.Fatalf("stale ordinary writer erased committed presence: preserved=%t error=%v", after.Equal(before), err)
	}
}
