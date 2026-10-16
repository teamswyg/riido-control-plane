package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestLifecycleACKProjectionWriterRejectsOversizedPrivateReceipt(t *testing.T) {
	ops, req := lifecycleACKFixture(t, AssignmentRunning)
	req.ProviderSessionID = strings.Repeat("native-session", 2048)
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer store.Close()
	if _, err := store.RecordAgentEvent(context.Background(), "agent", req); err == nil {
		t.Fatal("oversized private receipt committed")
	}
	if len(ops.records) != 0 || ops.projection.Assignment.State != AssignmentRunning || len(ops.projection.EventACKs) != 0 {
		t.Fatal("rejected receipt partially committed")
	}
	record := AssignmentOperationRecord{eventACK: &assignmentEventACKReceipt{ProviderSessionID: req.ProviderSessionID}}
	if _, err := assignmentProjectionDynamoDBItem(record); err == nil {
		t.Fatal("writer accepted an acknowledgement its decoder would reject")
	}
	if _, err := assignmentEventACKsFromJSON(strings.Repeat(" ", maxAssignmentEventACKBytes+1), 100); err == nil {
		t.Fatal("decoder accepted an oversized receipt")
	}
}
