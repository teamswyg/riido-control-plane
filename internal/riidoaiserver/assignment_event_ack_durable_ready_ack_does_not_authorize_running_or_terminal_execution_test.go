package riidoaiserver

import (
	"context"
	"reflect"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestDurableReadyACKDoesNotAuthorizeRunningOrTerminalExecution(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentLeased)
	req.State, req.EventType = AssignmentReady, EventAssignmentReady
	req.Metadata[metadatakeys.AssignmentEventKey.String()] = "ready:execution"
	req.Metadata["lease_token"] = "lease"
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer store.Close()
	ack, err := store.RecordAgentEvent(context.Background(), "agent", req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ack.Event.Metadata["lease_token"]; ok {
		t.Fatal("lease token leaked in Ready ACK")
	}
	replica := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer replica.Close()
	replay, err := replica.RecordAgentEvent(context.Background(), "agent", req)
	if err != nil || !reflect.DeepEqual(ack, replay) || len(ops.records) != 1 {
		t.Fatalf("active Ready replay changed: %+v %v", replay, err)
	}
	running := AgentEventRequest{AssignmentID: req.AssignmentID, State: AssignmentRunning, EventType: EventAssignmentRunning}
	if _, err := store.RecordAgentEvent(context.Background(), "agent", running); err != nil {
		t.Fatal(err)
	}
	if _, err := replica.RecordAgentEvent(context.Background(), "agent", req); err == nil {
		t.Fatal("already-running execution returned dispatch-authorizing Ready ACK")
	}
	terminal := AgentEventRequest{AssignmentID: req.AssignmentID, State: AssignmentCompleted, EventType: EventAssignmentCompleted, Metadata: map[string]string{metadatakeys.AssignmentEventKey.String(): "terminal"}}
	if _, err := store.RecordAgentEvent(context.Background(), "agent", terminal); err != nil {
		t.Fatal(err)
	}
	if _, err := replica.RecordAgentEvent(context.Background(), "agent", req); err == nil {
		t.Fatal("terminal execution returned dispatch-authorizing Ready ACK")
	}
	req.Metadata["lease_token"] = "wrong"
	if _, err := replica.RecordAgentEvent(context.Background(), "agent", req); err == nil {
		t.Fatal("wrong lease replay authorized")
	}
}
