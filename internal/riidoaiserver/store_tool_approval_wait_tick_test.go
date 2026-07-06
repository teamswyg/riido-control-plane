package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitForToolApprovalWithTickReturnsPendingAtHoldDeadline(t *testing.T) {
	store := NewStore()
	defer store.Close()

	pending := ToolApprovalResult{ApprovalID: "approval-pending", Status: ApprovalPending}
	result, decision, err := store.waitForToolApprovalWithTick(
		context.Background(), "agent-a", "asn-a", "approval-a",
		make(chan struct{}), 5*time.Millisecond, 0, pending, nil,
	)
	if err != nil {
		t.Fatalf("waitForToolApprovalWithTick: %v", err)
	}
	if decision != nil || result.Status != ApprovalPending {
		t.Fatalf("result = %+v, decision = %+v; want pending without decision", result, decision)
	}
}

func TestWaitForToolApprovalWithTickReturnsContextCancellation(t *testing.T) {
	store := NewStore()
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := store.waitForToolApprovalWithTick(
		ctx, "agent-a", "asn-a", "approval-a",
		make(chan struct{}), time.Second, time.Second, ToolApprovalResult{}, nil,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForToolApprovalWithTick error = %v, want context.Canceled", err)
	}
}
