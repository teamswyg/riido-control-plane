package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreTreatsStaleActiveRefreshAsIdempotent(t *testing.T) {
	fixedNow := time.Date(2026, 7, 7, 6, 10, 0, 0, time.UTC)
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		RequestBuffer: 1,
		Handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","message":"active lease changed"}`))
		},
	})

	assignment := sampleAssignmentOperationRecord(fixedNow).Assignment
	assignment.State = AssignmentRunning
	assignment.LeaseToken = "lease-1"
	err := store.RefreshAgentActiveAssignment(context.Background(), assignment, fixedNow)
	if err != nil {
		t.Fatalf("RefreshAgentActiveAssignment stale active lease: %v", err)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBUpdateItemTarget)
}

func TestDynamoDBAssignmentOperationStoreReportsActiveRefreshFailure(t *testing.T) {
	fixedNow := time.Date(2026, 7, 7, 6, 11, 0, 0, time.UTC)
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		RequestBuffer: 1,
		Handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"__type":"InternalServerError","Message":"active refresh unavailable"}`))
		},
	})

	assignment := sampleAssignmentOperationRecord(fixedNow).Assignment
	assignment.State = AssignmentRunning
	assignment.LeaseToken = "lease-1"
	err := store.RefreshAgentActiveAssignment(context.Background(), assignment, fixedNow)
	if err == nil || !strings.Contains(err.Error(), "active refresh unavailable") {
		t.Fatalf("RefreshAgentActiveAssignment error = %v", err)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBUpdateItemTarget)
}
