package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecoveryInventoryCancellationDoesNotBlockAssignmentActor(t *testing.T) {
	store, ops, client, req := preflightFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := store.RecordAgentEvent(ctx, ops.projection.Assignment.AgentID, req); done <- err }()
	<-client.entered
	metricsCtx, metricsCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer metricsCancel()
	if _, err := store.Metrics(metricsCtx); err != nil {
		t.Fatalf("inventory blocked actor: %v", err)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancel=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("inventory ignored caller cancellation")
	}
	if len(ops.records) != 0 {
		t.Fatal("cancelled preflight committed an event")
	}
}
