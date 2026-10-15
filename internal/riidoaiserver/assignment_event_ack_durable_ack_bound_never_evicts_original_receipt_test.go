package riidoaiserver

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestDurableACKBoundNeverEvictsOriginalReceipt(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentLeased)
	req.State, req.EventType = AssignmentReady, EventAssignmentReady
	req.Metadata[metadatakeys.AssignmentEventKey.String()] = "ready:original"
	req.Metadata["lease_token"] = "lease"
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer store.Close()
	first, err := store.RecordAgentEvent(context.Background(), "agent", req)
	if err != nil {
		t.Fatal(err)
	}
	original := req
	original.Metadata = cloneMetadata(req.Metadata)
	for i := 1; i < maxAssignmentEventACKReceipts; i++ {
		req.Metadata = cloneMetadata(req.Metadata)
		req.Metadata[metadatakeys.AssignmentEventKey.String()] = fmt.Sprintf("ready:attempt-%d", i)
		if _, err := store.RecordAgentEvent(context.Background(), "agent", req); err != nil {
			t.Fatal(err)
		}
	}
	req.Metadata = cloneMetadata(req.Metadata)
	req.Metadata[metadatakeys.AssignmentEventKey.String()] = "ready:over-budget"
	if _, err := store.RecordAgentEvent(context.Background(), "agent", req); err == nil {
		t.Fatal("receipt budget did not fail closed")
	}
	replica := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer replica.Close()
	again, err := replica.RecordAgentEvent(context.Background(), "agent", original)
	if err != nil || !reflect.DeepEqual(first, again) || len(ops.records) != maxAssignmentEventACKReceipts {
		t.Fatal("bound evicted original receipt or created duplicate")
	}
}
