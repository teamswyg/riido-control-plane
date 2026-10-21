package riidoaiserver

import (
	"context"
	"testing"
)

func TestRecoveryInventoryApprovalRechecksCASLeaseGeneration(t *testing.T) {
	store, ops, client, req := preflightFixture(t)
	close(client.release)
	ops.claimOnSave = true
	if _, err := store.RecordAgentEvent(context.Background(), ops.projection.Assignment.AgentID, req); err == nil {
		t.Fatal("approval survived another lease winning the CAS")
	}
	if len(ops.records) != 0 || ops.projection.Assignment.LeaseToken != "winner-lease" {
		t.Fatal("stale candidate approval committed")
	}
}
