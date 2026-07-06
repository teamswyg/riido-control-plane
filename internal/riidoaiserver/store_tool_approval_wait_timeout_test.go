package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestWaitForToolApprovalWithTickReturnsTimedOutApproval(t *testing.T) {
	ctx := context.Background()
	clock := &toolApprovalTestClock{now: time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)}
	store := NewStoreWithClock(clock.Now)
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-approval-timeout", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-approval",
		RuntimeProvider: "codex",
		Prompt:          "needs tool approval",
	})
	requestedAt := clock.Now()
	approval := mustCreateStoreToolApproval(t, store, ctx, ToolApprovalRequest{
		ApprovalID:   "approval-timeout",
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		AgentID:      assignment.AgentID,
		ToolID:       "tool-write",
		Status:       ApprovalPending,
		RequestedAt:  requestedAt,
		ExpiresAt:    requestedAt.Add(40 * time.Millisecond),
	})

	done := make(chan ToolApprovalResult, 1)
	go func() {
		result, _, err := store.WaitForToolApproval(
			ctx, assignment.AgentID, assignment.ID, approval.ApprovalID,
			200*time.Millisecond, 10*time.Millisecond,
		)
		if err != nil {
			t.Errorf("WaitForToolApproval: %v", err)
		}
		done <- result
	}()

	time.Sleep(60 * time.Millisecond)
	clock.Set(approval.ExpiresAt.Add(time.Millisecond))

	select {
	case result := <-done:
		if result.Status != ApprovalTimedOut {
			t.Fatalf("approval status = %s, want %s", result.Status, ApprovalTimedOut)
		}
		if !result.ResolvedAt.Equal(approval.ExpiresAt) {
			t.Fatalf("resolved_at = %s, want %s", result.ResolvedAt, approval.ExpiresAt)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForToolApproval did not return timed-out approval")
	}
}
