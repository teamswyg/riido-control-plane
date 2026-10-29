package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestAuthenticatedExecutionClaimDynamoTransaction(t *testing.T) {
	now := time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC)
	assignment := sampleQueuedAssignmentOperationRecord(now).Assignment
	assignment.CreatedBy = "requester"
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 1)
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now: now, RequestBuffer: 2, Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Amz-Target") == dynamoDBQueryTarget {
				_ = json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{item}})
			} else {
				_, _ = w.Write([]byte(`{}`))
			}
		},
	})
	ctx := context.WithValue(context.Background(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{deviceID: "device-a", ownerPrincipalID: "owner"})
	claim, err := store.ClaimNextAssignment(ctx, assignment.AgentID, now.Add(time.Second))
	if err != nil || !claim.Claimed {
		t.Fatalf("claim failed: %v", err)
	}
	<-requests
	var payload struct{ TransactItems []dynamoDBTransactWriteItem }
	if err := json.Unmarshal((<-requests).body, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.TransactItems) != 5 {
		t.Fatalf("authenticated claim requires atomic sidecar: actions=%d", len(payload.TransactItems))
	}
	if !isDevicePolicyAdvance(payload.TransactItems[4]) {
		t.Fatal("claim revision is not in the same CAS")
	}
	sidecar := payload.TransactItems[3].Put
	if sidecar.ConditionExpression != "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
		t.Fatal("claim sidecar is not immutable")
	}
	assertDynamoDBString(t, sidecar.Item, "pk", "DEVICE_EXECUTION#device-a")
	var persisted executionClaim
	if err := json.Unmarshal([]byte(dynamoDBStringValue(sidecar.Item, "payload")), &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.OwnerPrincipalID != "owner" || persisted.RequesterPrincipalID != "requester" || persisted.LeaseEpoch != leaseEpoch(claim.Assignment.LeaseToken) {
		t.Fatal("claim provenance mismatch")
	}
	if claim.Operation.ExecutionClaim == nil || *claim.Operation.ExecutionClaim != persisted {
		t.Fatal("journal lost claim provenance")
	}
}
