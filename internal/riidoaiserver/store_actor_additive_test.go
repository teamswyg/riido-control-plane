package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreActorAdditiveAssignmentKeepsExistingAgentActive(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	first := mustAssignActorTaskAdditive(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-1", RuntimeProvider: "codex", Prompt: "first",
	})
	firstPoll := mustPollActor(t, store, ctx, "agent-1")
	if firstPoll.Action != PollStart || firstPoll.Assignment == nil || firstPoll.Assignment.ID != first.ID {
		t.Fatalf("first poll = %+v", firstPoll)
	}

	now = now.Add(5 * time.Second)
	second := mustAssignActorTaskAdditive(t, store, ctx, "task-a", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-2", RuntimeProvider: "codex", Prompt: "second",
	})
	if second.ReplacesAssignmentID != "" || second.BlockedByAssignmentID != "" {
		t.Fatalf("additive assignment must not replace/block existing assignment: %+v", second)
	}

	firstPoll = mustPollActor(t, store, ctx, "agent-1")
	if firstPoll.Action != PollActive || firstPoll.Assignment == nil || firstPoll.Assignment.ID != first.ID {
		t.Fatalf("first active poll = %+v", firstPoll)
	}
	secondPoll := mustPollActor(t, store, ctx, "agent-2")
	if secondPoll.Action != PollStart || secondPoll.Assignment == nil || secondPoll.Assignment.ID != second.ID {
		t.Fatalf("second poll = %+v", secondPoll)
	}
}
