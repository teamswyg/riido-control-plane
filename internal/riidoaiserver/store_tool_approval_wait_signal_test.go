package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestWaitForToolApprovalWithTickReturnsApprovedSignal(t *testing.T) {
	ctx := context.Background()
	clock := &toolApprovalTestClock{now: time.Date(2026, 7, 7, 13, 0, 0, 0, time.UTC)}
	store := NewStoreWithClock(clock.Now)
	defer store.Close()

	assignment := mustAssignActorTask(t, store, ctx, "task-approval-signal", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-approval",
		RuntimeProvider: "claude",
		Prompt:          "needs write approval",
	})
	approval := mustCreateStoreToolApproval(t, store, ctx, ToolApprovalRequest{
		ApprovalID:   "approval-signal",
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		AgentID:      assignment.AgentID,
		ToolID:       "tool-write",
		Status:       ApprovalPending,
		RequestedAt:  clock.Now(),
		ExpiresAt:    clock.Now().Add(time.Second),
	})

	done := make(chan ToolApprovalResult, 1)
	go func() {
		result, _, err := store.WaitForToolApproval(
			ctx, assignment.AgentID, assignment.ID, approval.ApprovalID,
			200*time.Millisecond, 20*time.Millisecond,
		)
		if err != nil {
			t.Errorf("WaitForToolApproval: %v", err)
		}
		done <- result
	}()

	time.Sleep(20 * time.Millisecond)
	mustDecideStoreToolApproval(t, store, ctx, assignment.TaskID, approval, ApprovalDecisionApprove)
	select {
	case result := <-done:
		if result.Status != ApprovalApproved {
			t.Fatalf("approval status = %s, want %s", result.Status, ApprovalApproved)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForToolApproval did not return approved approval")
	}
}
