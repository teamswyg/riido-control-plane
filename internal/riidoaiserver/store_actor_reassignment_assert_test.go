package riidoaiserver

import (
	"context"
	"testing"
)

func assertSameBlockedQueuedAssignmentRepair(
	t *testing.T,
	store *Store,
	ctx context.Context,
	current AssignmentOperationRecord,
	blocker AssignmentOperationRecord,
) {
	t.Helper()
	reassigned := mustAssignActorTask(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-new", RuntimeProvider: "codex", Prompt: "second",
	})
	if reassigned.ID != current.Assignment.ID || reassigned.BlockedByAssignmentID != "" {
		t.Fatalf("reassigned current = %+v", reassigned)
	}
	if projection := mustLoadActorProjection(t, store, ctx, blocker.Assignment.ID); projection.Assignment.State != AssignmentCancelled {
		t.Fatalf("blocker projection = %+v", projection)
	}
	currentProjection := mustLoadActorProjection(t, store, ctx, current.Assignment.ID)
	if currentProjection.Assignment.BlockedByAssignmentID != "" {
		t.Fatalf("current projection = %+v", currentProjection)
	}
	poll := mustPollActor(t, store, ctx, "agent-new")
	if poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != current.Assignment.ID {
		t.Fatalf("poll after blocker repair = %+v", poll)
	}
}
