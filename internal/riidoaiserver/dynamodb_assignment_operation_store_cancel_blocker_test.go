package riidoaiserver

import (
	"testing"
	"time"
)

func TestDynamoDBCancelQueuedBlockerOperationCancelsBlocker(t *testing.T) {
	queuedAt := time.Date(2026, 7, 7, 4, 30, 0, 0, time.UTC)
	cancelAt := queuedAt.Add(3 * time.Second)
	blocker := sampleQueuedAssignmentOperationRecord(queuedAt).Assignment
	blocker.ID = "asn-blocker"
	blockedID := "asn-blocked"

	operation := dynamoDBCancelQueuedBlockerOperation(
		assignmentProjectionRecord{Assignment: blocker, LastEventSeq: 7},
		blockedID,
		cancelAt,
	)

	if operation.OperationType != AssignmentOperationAgentEvent {
		t.Fatalf("operation type = %q", operation.OperationType)
	}
	if operation.Assignment.State != AssignmentCancelled {
		t.Fatalf("assignment state = %q", operation.Assignment.State)
	}
	if operation.Assignment.UpdatedAt != cancelAt {
		t.Fatalf("assignment updated_at = %s", operation.Assignment.UpdatedAt)
	}
	if len(operation.Events) != 1 {
		t.Fatalf("events = %+v", operation.Events)
	}
	event := operation.Events[0]
	if event.Seq != 8 || event.Type != EventAssignmentCancelled || event.State != AssignmentCancelled {
		t.Fatalf("event = %+v", event)
	}
	if event.Message != "queued blocker was cancelled before queued assignment claim" {
		t.Fatalf("event message = %q", event.Message)
	}
	if event.Metadata["blocked_assignment_id"] != blockedID {
		t.Fatalf("event metadata = %+v", event.Metadata)
	}
}
