package riidoaiserver

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestToolApprovalWithoutDeadlineWaitsForExplicitDecision(t *testing.T) {
	ctx := context.Background()
	clock := &toolApprovalTestClock{now: time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)}
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	store, err := OpenStoreWithConfig(ctx, StoreConfig{SnapshotStore: snapshot, Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	assignment := mustAssignActorTask(t, store, ctx, "task-human-wait", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-approval", RuntimeProvider: "codex", Prompt: "needs approval",
	})
	approval := mustCreateStoreToolApproval(t, store, ctx, ToolApprovalRequest{
		ApprovalID: "approval-1", AssignmentID: assignment.ID, TaskID: assignment.TaskID,
		AgentID: assignment.AgentID, ToolID: "tool-1", ToolName: "headroom_retrieve",
	})
	store.Close()
	store, err = OpenStoreWithConfig(ctx, StoreConfig{SnapshotStore: snapshot, Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	clock.Set(clock.Now().Add(24 * time.Hour))
	result, decision, err := store.WaitForToolApproval(ctx, assignment.AgentID, assignment.ID, approval.ApprovalID, 0, 0)
	if err != nil || decision != nil || result.Status != ApprovalPending {
		t.Fatalf("human request expired without a decision: result=%+v decision=%+v err=%v", result, decision, err)
	}
	approvals, err := store.ListTaskToolApprovals(ctx, assignment.TaskID)
	if err != nil || len(approvals) != 1 || approvals[0].ToolName != "headroom_retrieve" || approvals[0].Status != ApprovalPending {
		t.Fatalf("approval details not preserved: %+v err=%v", approvals, err)
	}
	mustDecideStoreToolApproval(t, store, ctx, assignment.TaskID, approval, ApprovalDecisionApprove)
}

func TestToolApprovalAfterCancellationCannotResumeOldRequest(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()
	assignment := mustAssignActorTask(t, store, ctx, "task-cancel", AssignRequest{
		ComponentID: "component-a", AgentID: "agent-approval", RuntimeProvider: "codex", Prompt: "needs approval",
	})
	approval := mustCreateStoreToolApproval(t, store, ctx, ToolApprovalRequest{
		ApprovalID: "approval-1", AssignmentID: assignment.ID, TaskID: assignment.TaskID,
		AgentID: assignment.AgentID, ToolID: "tool-1",
	})
	if _, err := store.CancelAssignment(ctx, assignment.TaskID, CancelAssignmentRequest{AgentID: assignment.AgentID, AssignmentID: assignment.ID}); err != nil {
		t.Fatal(err)
	}
	_, decision, err := store.DecideToolApproval(ctx, assignment.TaskID, ToolApprovalDecision{
		AssignmentID: assignment.ID, ApprovalID: approval.ApprovalID, Decision: ApprovalDecisionApprove, DecidedBy: "user",
	})
	if err == nil || decision != nil {
		t.Fatalf("cancelled request accepted: decision=%+v err=%v", decision, err)
	}
}
