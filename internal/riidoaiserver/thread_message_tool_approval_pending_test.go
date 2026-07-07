package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPendingThreadToolApproval(t *testing.T) {
	now := time.Now().UTC().Add(time.Minute)
	thread := AIAgentTaskThreadRecord{AssignmentID: "asn-1", AgentID: "agent-1"}
	pending := ToolApprovalRequest{
		ApprovalID:   "approval-1",
		TaskID:       "task-1",
		AssignmentID: "asn-1",
		AgentID:      "agent-1",
		Status:       ApprovalPending,
		ExpiresAt:    now,
	}
	t.Run("store error fails closed", func(t *testing.T) {
		store := pendingToolApprovalStore{err: errors.New("boom")}
		_, ok, err := pendingThreadToolApproval(context.Background(), store, "task-1", thread)
		if err == nil || ok {
			t.Fatalf("ok=%v err=%v, want error", ok, err)
		}
	})
	t.Run("no matching approval", func(t *testing.T) {
		store := pendingToolApprovalStore{approvals: []ToolApprovalRequest{{
			ApprovalID:   "approval-2",
			AssignmentID: "asn-2",
			AgentID:      "agent-1",
			Status:       ApprovalPending,
		}}}
		_, ok, err := pendingThreadToolApproval(context.Background(), store, "task-1", thread)
		if err != nil || ok {
			t.Fatalf("ok=%v err=%v, want no match", ok, err)
		}
	})
	t.Run("returns first matching pending approval", func(t *testing.T) {
		store := pendingToolApprovalStore{approvals: []ToolApprovalRequest{pending}}
		got, ok, err := pendingThreadToolApproval(context.Background(), store, "task-1", thread)
		if err != nil || !ok || got.ApprovalID != pending.ApprovalID {
			t.Fatalf("got=%+v ok=%v err=%v", got, ok, err)
		}
	})
}

type pendingToolApprovalStore struct {
	approvals []ToolApprovalRequest
	err       error
}

func (s pendingToolApprovalStore) ListTaskToolApprovals(context.Context, string) ([]ToolApprovalRequest, error) {
	return s.approvals, s.err
}

func (pendingToolApprovalStore) CreateToolApproval(context.Context, string, ToolApprovalRequest) (ToolApprovalRequest, error) {
	return ToolApprovalRequest{}, nil
}

func (pendingToolApprovalStore) DecideToolApproval(context.Context, string, ToolApprovalDecision) (ToolApprovalResult, *ToolApprovalDecision, error) {
	return ToolApprovalResult{}, nil, nil
}

func (pendingToolApprovalStore) WaitForToolApproval(context.Context, string, string, string, time.Duration, time.Duration) (ToolApprovalResult, *ToolApprovalDecision, error) {
	return ToolApprovalResult{}, nil, nil
}
