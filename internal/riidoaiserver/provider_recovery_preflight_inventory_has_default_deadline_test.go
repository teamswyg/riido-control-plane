package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecoveryInventoryHasDefaultDeadline(t *testing.T) {
	store, ops, client, req := preflightFixture(t)
	started := time.Now()
	_, err := store.RecordAgentEvent(context.Background(), ops.projection.Assignment.AgentID, req)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 3*time.Second {
		t.Fatalf("unbounded inventory: %v elapsed=%v", err, time.Since(started))
	}
	select {
	case <-client.entered:
	default:
		t.Fatal("test did not enter lookup")
	}
	if len(ops.records) != 0 {
		t.Fatal("expired preflight committed an event")
	}
}
