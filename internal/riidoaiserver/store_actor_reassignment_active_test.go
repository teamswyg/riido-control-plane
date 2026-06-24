package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreActorReassignmentCancelsPreviousAndBlocksNewAgent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	first := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "first",
	})
	mustPollActor(t, store, ctx, "agent-1")

	now = now.Add(5 * time.Second)
	second := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-2", RuntimeProvider: "codex", Prompt: "second",
	})
	if second.ReplacesAssignmentID != first.ID || second.BlockedByAssignmentID != first.ID || second.State != AssignmentQueued {
		t.Fatalf("second assignment = %+v", second)
	}
	if pollSecond := mustPollActor(t, store, ctx, "agent-2"); pollSecond.Action != PollNone {
		t.Fatalf("poll second before cancel = %+v", pollSecond)
	}
	pollFirst := mustPollActor(t, store, ctx, "agent-1")
	if pollFirst.Action != PollCancel || pollFirst.Assignment == nil || pollFirst.Assignment.ID != first.ID {
		t.Fatalf("poll first cancel = %+v", pollFirst)
	}

	mustRecordActorEvent(t, store, ctx, "agent-1", AgentEventRequest{
		AssignmentID: first.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentCancelled,
		EventType:    EventAssignmentCancelled,
	})
	pollSecond := mustPollActor(t, store, ctx, "agent-2")
	if pollSecond.Action != PollStart || pollSecond.Assignment == nil || pollSecond.Assignment.ID != second.ID {
		t.Fatalf("poll second start = %+v", pollSecond)
	}
}
