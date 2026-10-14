package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreActorReassignmentCancelsQueuedPreviousWithoutBlockingNewAgent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	first := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "first",
	})
	now = now.Add(time.Second)
	second := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-2", RuntimeProvider: "codex", Prompt: "second",
	})
	if second.ReplacesAssignmentID != first.ID || second.BlockedByAssignmentID != "" {
		t.Fatalf("second assignment should replace without blocker: %+v", second)
	}
	if firstProjection := mustLoadActorProjection(t, store, ctx, first.ID); firstProjection.Assignment.State != AssignmentCancelled {
		t.Fatalf("first projection = %+v", firstProjection)
	}
	secondPoll := mustPollActor(t, store, ctx, "agent-2")
	if secondPoll.Action != PollStart || secondPoll.Assignment == nil || secondPoll.Assignment.ID != second.ID {
		t.Fatalf("second poll = %+v", secondPoll)
	}
	if firstPoll := mustPollActor(t, store, ctx, "agent-1"); firstPoll.Action != PollNone {
		t.Fatalf("first poll after queued cancellation = %+v", firstPoll)
	}
}

func TestStoreActorReassigningSameBlockedQueuedAssignmentRepairsQueuedBlocker(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	blocker := replayOperationRecord("task-a", "asn-000020", "agent-old", AssignmentQueued, 1, base)
	current := replayOperationRecord("task-a", "asn-000027", "agent-new", AssignmentQueued, 2, base.Add(time.Second))
	current.Assignment.BlockedByAssignmentID = blocker.Assignment.ID
	current.Assignment.ReplacesAssignmentID = blocker.Assignment.ID
	current.OperationID = assignmentOperationID(current.OperationType, current.Assignment, current.Events)
	operations := &runtimeFakeAssignmentOperationStore{records: []AssignmentOperationRecord{blocker, current}}
	now := base.Add(2 * time.Second)
	store, err := OpenStoreWithConfig(ctx, StoreConfig{Now: func() time.Time { return now }, OperationStore: operations})
	if err != nil {
		t.Fatalf("OpenStoreWithConfig: %v", err)
	}
	defer store.Close()
	assertSameBlockedQueuedAssignmentRepair(t, store, ctx, current, blocker)
}
