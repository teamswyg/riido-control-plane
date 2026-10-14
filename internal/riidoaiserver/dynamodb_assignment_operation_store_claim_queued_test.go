package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreClaimsQueuedAssignmentWithTransaction(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	claimAt := fixedNow.Add(time.Second)
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
	if !claim.Claimed || claim.Assignment.ID != assignment.ID || claim.Assignment.State != AssignmentLeased {
		t.Fatalf("claim = %+v", claim)
	}
	wantLeaseToken := "asn-000001:" + strconv.FormatInt(claimAt.UnixNano(), 10)
	if claim.Assignment.LeaseToken != wantLeaseToken {
		t.Fatalf("lease token = %q", claim.Assignment.LeaseToken)
	}
	if claim.Operation.OperationType != AssignmentOperationPollStart || len(claim.Operation.Events) != 1 {
		t.Fatalf("operation = %+v", claim.Operation)
	}
	if claim.Operation.Events[0].Seq != 2 || claim.Operation.Events[0].Type != EventAssignmentLeased {
		t.Fatalf("claim event = %+v", claim.Operation.Events[0])
	}

	queryRequest := <-requests
	assertDynamoDBTarget(t, queryRequest, dynamoDBQueryTarget)
	transactRequest := <-requests
	assertDynamoDBTarget(t, transactRequest, dynamoDBTransactWriteTarget)
	payload := decodeDynamoDBTransactWritePayload(t, transactRequest)
	assertQueuedClaimTransactPayload(t, claim, claimAt, payload)
}
