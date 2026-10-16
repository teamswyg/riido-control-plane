package riidoaiserver

import (
	"context"
	"testing"
)

func TestAgentEventMissingTargetStaysValidationBeforeDependencyLookup(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentLeased)
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	t.Cleanup(store.Close)
	for _, tc := range []struct {
		agentID      string
		assignmentID string
	}{{"", req.AssignmentID}, {"agent", "   "}} {
		req.AssignmentID = tc.assignmentID
		_, err := store.RecordAgentEvent(context.Background(), tc.agentID, req)
		if err == nil {
			t.Fatal("missing target was accepted")
		}
		if status, _ := agentEventErrorResponse(err); status != 400 {
			t.Fatal("missing target was misclassified as a retryable dependency")
		}
	}
}
