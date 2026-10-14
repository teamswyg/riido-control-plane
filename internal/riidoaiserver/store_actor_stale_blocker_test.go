package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreActorPollRepairsStaleBlockedQueuedAssignment(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 9, 4, 0, 0, 0, time.UTC)
	operations := &runtimeFakeActiveLeaseOperationStore{}
	store := NewStoreWithConfig(StoreConfig{
		Now:                 func() time.Time { return now },
		ActiveLeaseDuration: time.Minute,
		OperationStore:      operations,
	})
	defer store.Close()

	first, firstPoll := assignAndStartStaleBlocker(t, store, ctx)
	operations.activeFound = true
	operations.activeLease = AssignmentActiveLease{
		AgentID:            "agent-old",
		ActiveAssignmentID: first.ID,
		LeaseToken:         firstPoll.Assignment.LeaseToken,
		HeartbeatAt:        now,
		LeaseExpiresAt:     now.Add(time.Minute),
		LeaseExpiresUnixMS: now.Add(time.Minute).UnixMilli(),
	}

	now = now.Add(10 * time.Second)
	second := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-new", RuntimeProvider: "codex", Prompt: "second",
	})
	if second.ReplacesAssignmentID != first.ID || second.BlockedByAssignmentID != first.ID {
		t.Fatalf("second assignment = %+v", second)
	}
	if beforeExpiry := mustPollActor(t, store, ctx, "agent-new"); beforeExpiry.Action != PollNone {
		t.Fatalf("new assignment should stay blocked before lease expiry: %+v", beforeExpiry)
	}

	now = now.Add(2 * time.Minute)
	assertStaleBlockedAssignmentRepaired(t, store, ctx, second)
	if oldPollAfterRepair := mustPollActor(t, store, ctx, "agent-old"); oldPollAfterRepair.Action != PollNone {
		t.Fatalf("old assignment should not stay active after stale blocker repair: %+v", oldPollAfterRepair)
	}
	assertStaleBlockerRepairOperations(t, operations.records, first.ID, second.ID)
}
