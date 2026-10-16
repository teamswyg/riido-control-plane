package riidoaiserver

import (
	"context"
	"reflect"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestReadyACKPreservesLegacyV88KeyWithoutLeaseMetadata(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentLeased)
	req.State, req.EventType = AssignmentReady, EventAssignmentReady
	req.Metadata[metadatakeys.AssignmentEventKey.String()] = "daemon:assignment:1"
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer store.Close()
	ack, err := store.RecordAgentEvent(context.Background(), "agent", req)
	if err != nil || ack.Assignment.State != AssignmentReady {
		t.Fatalf("legacy Ready rejected: %v", err)
	}
	replica := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer replica.Close()
	again, err := replica.RecordAgentEvent(context.Background(), "agent", req)
	if err != nil || !reflect.DeepEqual(ack, again) || len(ops.records) != 1 {
		t.Fatal("legacy Ready replay changed")
	}
}
