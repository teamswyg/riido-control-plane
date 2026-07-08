package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDecideToolApprovalAfterExpiryReturnsTimedOutWithoutDecision(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	clock := &toolApprovalTestClock{now: now}
	store := NewStoreWithClock(clock.Now)
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-expired-approval", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-approval",
		RuntimeProvider: "claude",
		Prompt:          "needs file write approval",
	})
	approval := mustCreateStoreToolApproval(t, store, ctx, ToolApprovalRequest{
		ApprovalID:   "approval-expired",
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		AgentID:      assignment.AgentID,
		ToolID:       "tool-write",
		Status:       ApprovalPending,
		RequestedAt:  now,
		ExpiresAt:    now.Add(time.Minute),
	})
	clock.Set(approval.ExpiresAt.Add(time.Second))

	result, decision, err := store.DecideToolApproval(ctx, assignment.TaskID, ToolApprovalDecision{
		ApprovalID:   approval.ApprovalID,
		AssignmentID: approval.AssignmentID,
		Decision:     ApprovalDecisionApprove,
		DecidedBy:    "late-user",
	})
	if err != nil {
		t.Fatalf("DecideToolApproval: %v", err)
	}
	if decision != nil {
		t.Fatalf("decision = %+v, want nil for expired approval", decision)
	}
	if result.Status != ApprovalTimedOut || !result.ResolvedAt.Equal(approval.ExpiresAt) {
		t.Fatalf("result = %+v, want timed out at %s", result, approval.ExpiresAt)
	}
}
