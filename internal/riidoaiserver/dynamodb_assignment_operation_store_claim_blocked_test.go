package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreSkipsBlockedAssignmentUntilBlockerTerminal(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	assignment := sampleQueuedAssignmentOperationRecord(fixedNow).Assignment
	assignment.BlockedByAssignmentID = "asn-000000"
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 1)
	blocker := Assignment{
		ID:              "asn-000000",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim-old",
		RuntimeProvider: "codex",
		Prompt:          "old work",
		State:           AssignmentRunning,
		CreatedAt:       fixedNow,
		UpdatedAt:       fixedNow,
	}
	blockerItem := sampleAssignmentProjectionDynamoDBItem(t, blocker, 9)
	activeItem := blockedClaimActiveItem(fixedNow)
	getCalls := 0
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		RequestBuffer: 3,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			switch r.Header.Get("X-Amz-Target") {
			case dynamoDBQueryTarget:
				if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{item}}); err != nil {
					t.Errorf("encode query response: %v", err)
				}
			case dynamoDBGetItemTarget:
				getCalls++
				responseItem := blockerItem
				if getCalls == 2 {
					responseItem = activeItem
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"Item": responseItem}); err != nil {
					t.Errorf("encode blocker response: %v", err)
				}
			case dynamoDBTransactWriteTarget:
				t.Errorf("blocked assignment must not be transaction-claimed")
				w.WriteHeader(http.StatusBadRequest)
			default:
				t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
				w.WriteHeader(http.StatusBadRequest)
			}
		},
	})

	claim, err := store.ClaimNextAssignment(context.Background(), "jykim1", fixedNow)
	if err != nil {
		t.Fatalf("ClaimNextAssignment blocked: %v", err)
	}
	if claim.Claimed {
		t.Fatalf("blocked assignment should not be claimed: %+v", claim)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBQueryTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBGetItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBGetItemTarget)
}
