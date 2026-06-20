package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestDynamoDBAssignmentOperationStoreRepairsStaleQueuedInsteadOfClaim(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	claimAt := fixedNow.Add(staleReplayQueuedAssignmentMaxAge + time.Minute)
	assignment := sampleQueuedAssignmentOperationRecord(fixedNow).Assignment
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 1)
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		RequestBuffer: 2,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			switch r.Header.Get("X-Amz-Target") {
			case dynamoDBQueryTarget:
				if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{item}}); err != nil {
					t.Errorf("encode query response: %v", err)
				}
			case dynamoDBTransactWriteTarget:
				_, _ = w.Write([]byte(`{}`))
			default:
				t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
				w.WriteHeader(http.StatusBadRequest)
			}
		},
	})

	claim, err := store.ClaimNextAssignment(context.Background(), "jykim1", claimAt)
	if err != nil {
		t.Fatalf("ClaimNextAssignment: %v", err)
	}
	if claim.Claimed {
		t.Fatalf("stale queued assignment must not be claimed: %+v", claim)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBQueryTarget)
	payload := decodeDynamoDBTransactWritePayload(t, <-requests)
	if len(payload.TransactItems) != 2 {
		t.Fatalf("transact item count = %d", len(payload.TransactItems))
	}
	projectionPut := payload.TransactItems[0].Put
	assertDynamoDBString(t, projectionPut.Item, "pk", dynamoDBAssignmentProjectionPK(assignment.ID))
	assertDynamoDBString(t, projectionPut.Item, "assignment_state", string(AssignmentFailed))
	assertDynamoDBString(t, projectionPut.ExpressionAttributeValues, ":expected_state", string(AssignmentQueued))
	assertDynamoDBNumber(t, projectionPut.ExpressionAttributeValues, ":expected_last_event_seq", "1")
	if _, ok := projectionPut.Item["agent_id"]; ok {
		t.Fatalf("failed projection must leave agent_queue: %+v", projectionPut.Item)
	}

	operationPut := payload.TransactItems[1].Put
	assertDynamoDBString(t, operationPut.Item, "operation_type", string(AssignmentOperationAgentEvent))
	assertDynamoDBString(t, operationPut.Item, "assignment_state", string(AssignmentFailed))
	assertDynamoDBNumber(t, operationPut.Item, "last_event_seq", "2")
	var events []TaskEvent
	if err := json.Unmarshal([]byte(operationPut.Item["events_json"]["S"]), &events); err != nil {
		t.Fatalf("decode events: %v", err)
	}
	categoryKey := metadatakeys.AssignmentFailureCategory.String()
	if len(events) != 1 || events[0].Metadata[categoryKey] != "stale_queued_assignment" {
		t.Fatalf("events = %+v", events)
	}
}
