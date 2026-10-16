package riidoaiserver

import (
	"context"
	"testing"
)

func TestEventOnlyTerminalTypeDoesNotCorruptACKProjection(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentRunning)
	req.State = ""
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer store.Close()
	if _, err := store.RecordAgentEvent(context.Background(), "agent", req); err != nil {
		t.Fatal(err)
	}
	if len(ops.projection.EventACKs) != 0 {
		t.Fatal("nonterminal event retained a terminal receipt")
	}
	item, err := assignmentProjectionDynamoDBItem(ops.records[0])
	if err != nil {
		t.Fatal(err)
	}
	if _, err := assignmentProjectionFromDynamoDBItem(item); err != nil {
		t.Fatalf("writer produced unreadable projection: %v", err)
	}
}
