package riidoaiserver

import (
	"context"
	"reflect"
	"testing"
)

func TestDurableACKRefusesChangedPayloadWithoutReplacingReceipt(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentRunning)
	first := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer first.Close()
	ack, err := first.RecordAgentEvent(context.Background(), "agent", req)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []string{"message", "metadata", "type", "state", "session", "identity"} {
		t.Run(change, func(t *testing.T) {
			next := req
			next.Metadata = cloneMetadata(req.Metadata)
			switch change {
			case "message":
				next.Message = "different"
			case "metadata":
				next.Metadata["extra"] = "value"
			case "type":
				next.EventType = EventAssignmentCompleted
			case "state":
				next.State = AssignmentCompleted
			case "session":
				next.ProviderSessionID = "foreign"
			case "identity":
				next.RuntimeID = "foreign"
			}
			replica := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer replica.Close()
			if _, err := replica.RecordAgentEvent(context.Background(), "agent", next); err == nil {
				t.Fatal("conflicting payload was acknowledged")
			}
			original, err := replica.RecordAgentEvent(context.Background(), "agent", req)
			if err != nil || !reflect.DeepEqual(original, ack) || len(ops.records) != 1 {
				t.Fatal("rejected reuse replaced receipt or emitted another event")
			}
		})
	}
}
