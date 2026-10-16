package riidoaiserver

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestDurableReadyACKRejectsExpiredLease(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentLeased)
	req.State, req.EventType = AssignmentReady, EventAssignmentReady
	req.Metadata[metadatakeys.AssignmentEventKey.String()] = "ready:expiry"
	req.Metadata["lease_token"] = "lease"
	initial := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer initial.Close()
	if _, err := initial.RecordAgentEvent(context.Background(), "agent", req); err != nil {
		t.Fatal(err)
	}
	ops.projection.Assignment.UpdatedAt = time.Now().Add(-time.Minute)
	replica := NewStoreWithConfig(StoreConfig{OperationStore: &expiredACKLeaseOperations{ops}})
	defer replica.Close()
	if _, err := replica.RecordAgentEvent(context.Background(), "agent", req); err == nil || len(ops.records) != 1 {
		t.Fatal("expired Ready ACK granted dispatch")
	}
}
