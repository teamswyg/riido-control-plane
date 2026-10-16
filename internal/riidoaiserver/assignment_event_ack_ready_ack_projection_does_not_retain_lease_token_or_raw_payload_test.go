package riidoaiserver

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestReadyACKProjectionDoesNotRetainLeaseTokenOrRawPayload(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentLeased)
	req.State, req.EventType = AssignmentReady, EventAssignmentReady
	req.Metadata[metadatakeys.AssignmentEventKey.String()] = "ready:private-key"
	req.Metadata["lease_token"] = "lease"
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer store.Close()
	ack, err := store.RecordAgentEvent(context.Background(), "agent", req)
	if err != nil {
		t.Fatal(err)
	}
	item, err := assignmentProjectionDynamoDBItem(ops.records[0])
	if err != nil {
		t.Fatal(err)
	}
	if ackContainsAny(item["event_ack_receipts_json"]["S"], "ready:private-key", req.Message, "lease_token", `"lease"`) {
		t.Fatal("private receipt retained raw request payload or credential")
	}
	if _, ok := ack.Event.Metadata["lease_token"]; ok {
		t.Fatal("Ready ACK retained credential")
	}
	if _, ok := ops.records[0].Events[0].Metadata["lease_token"]; ok {
		t.Fatal("operation event retained credential")
	}
}
