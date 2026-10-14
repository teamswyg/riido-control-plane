package riidoaiserver

import (
	"context"
	"sync"
	"testing"
	"time"
)

func mustCreateStoreToolApproval(t *testing.T, store *Store, ctx context.Context, req ToolApprovalRequest) ToolApprovalRequest {
	t.Helper()
	approval, err := store.CreateToolApproval(ctx, req.AgentID, req)
	if err != nil {
		t.Fatalf("CreateToolApproval: %v", err)
	}
	return approval
}

func mustDecideStoreToolApproval(t *testing.T, store *Store, ctx context.Context, taskID string, approval ToolApprovalRequest, decision ApprovalDecision) {
	t.Helper()
	result, saved, err := store.DecideToolApproval(ctx, taskID, ToolApprovalDecision{
		ApprovalID:   approval.ApprovalID,
		AssignmentID: approval.AssignmentID,
		Decision:     decision,
		DecidedBy:    "test-user",
	})
	if err != nil {
		t.Fatalf("DecideToolApproval: %v", err)
	}
	want := ApprovalDenied
	if decision == ApprovalDecisionApprove {
		want = ApprovalApproved
	}
	if saved == nil || result.Status != want {
		t.Fatalf("tool approval decision = %+v, result = %+v", saved, result)
	}
}

type toolApprovalTestClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *toolApprovalTestClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *toolApprovalTestClock) Set(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = now
}
