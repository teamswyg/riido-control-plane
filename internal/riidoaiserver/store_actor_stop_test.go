package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreActorClientStopCancelsActiveAssignmentForDaemonPoll(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 13, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "first",
	})
	pollStart := mustPollActor(t, store, ctx, "agent-1")
	if pollStart.Action != PollStart || pollStart.Assignment == nil || pollStart.Assignment.ID != assignment.ID {
		t.Fatalf("poll start = %+v", pollStart)
	}

	now = now.Add(time.Second)
	cancelled := mustCancelActorAssignment(t, store, ctx, "task-a", CancelAssignmentRequest{
		AgentID: "agent-1", AssignmentID: assignment.ID, Reason: "user requested stop",
	})
	if cancelled.State != AssignmentCancelling {
		t.Fatalf("cancelled assignment = %+v", cancelled)
	}
	if projection := mustLoadActorProjection(t, store, ctx, assignment.ID); projection.Assignment.State != AssignmentCancelling {
		t.Fatalf("projection after cancel = %+v", projection)
	}
	pollCancel := mustPollActor(t, store, ctx, "agent-1")
	if pollCancel.Action != PollCancel || pollCancel.Assignment == nil || pollCancel.Assignment.ID != assignment.ID {
		t.Fatalf("poll cancel = %+v", pollCancel)
	}
}

func TestStoreActorClientStopCancelsQueuedAssignmentImmediately(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 14, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "first",
	})
	cancelled := mustCancelActorAssignment(t, store, ctx, "task-a", CancelAssignmentRequest{
		AgentID: "agent-1", AssignmentID: assignment.ID, Reason: "user requested stop",
	})
	if cancelled.State != AssignmentCancelled {
		t.Fatalf("cancelled queued assignment = %+v", cancelled)
	}
	if poll := mustPollActor(t, store, ctx, "agent-1"); poll.Action != PollNone {
		t.Fatalf("poll after queued cancel = %+v", poll)
	}
}
