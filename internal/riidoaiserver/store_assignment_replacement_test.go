package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestStoreAssignTaskReplacementCreatesNewAssignmentForSameAgent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()

	first, err := store.AssignTask(ctx, "task-repeat", replacementTestAssignRequest("first"))
	if err != nil {
		t.Fatalf("AssignTask first: %v", err)
	}
	now = now.Add(time.Second)
	second, err := store.AssignTaskReplacement(ctx, "task-repeat", replacementTestAssignRequest("second"))
	if err != nil {
		t.Fatalf("AssignTaskReplacement second: %v", err)
	}
	if second.ID == first.ID || second.ReplacesAssignmentID != first.ID {
		t.Fatalf("replacement assignment = %+v first=%+v", second, first)
	}
	firstProjection, ok, err := store.LoadAssignmentProjection(ctx, first.ID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection first: %v", err)
	}
	if !ok || firstProjection.Assignment.State != AssignmentCancelled {
		t.Fatalf("first projection = %+v ok=%v", firstProjection, ok)
	}
}

func replacementTestAssignRequest(prompt string) AssignRequest {
	return AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          prompt,
	}
}
