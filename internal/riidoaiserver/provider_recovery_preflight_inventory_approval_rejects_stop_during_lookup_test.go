package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestRecoveryInventoryApprovalRejectsStopDuringLookup(t *testing.T) {
	store, ops, client, req := preflightFixture(t)
	done := make(chan error, 1)
	go func() {
		_, err := store.RecordAgentEvent(context.Background(), ops.projection.Assignment.AgentID, req)
		done <- err
	}()
	<-client.entered
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := store.CancelAssignment(ctx, "task", CancelAssignmentRequest{AgentID: ops.projection.Assignment.AgentID, AssignmentID: req.AssignmentID}); err != nil {
		t.Fatal(err)
	}
	close(client.release)
	if err := <-done; err == nil {
		t.Fatal("stale preflight authorized dispatch after stop")
	}
	if len(ops.records) != 1 || ops.projection.Assignment.State != AssignmentCancelling {
		t.Fatalf("stop lost: %+v", ops.projection)
	}
}
