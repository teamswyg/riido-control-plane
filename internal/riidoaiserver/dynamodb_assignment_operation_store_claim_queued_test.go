package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreClaimsQueuedAssignmentWithTransaction(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	claimAt := fixedNow.Add(time.Second)
	requests := make(chan capturedDynamoDBRequest, 2)
	assignment := sampleQueuedAssignmentOperationRecord(fixedNow).Assignment
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{
			method: r.Method,
			header: r.Header.Clone(),
			body:   body,
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
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
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAssignmentOperationStore: %v", err)
	}
	defer store.Close()

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
	if len(payload.TransactItems) != 3 {
		t.Fatalf("transact item count = %d", len(payload.TransactItems))
	}
	projectionPut := payload.TransactItems[0].Put
	if projectionPut.ConditionExpression != "assignment_state = :queued AND agent_id = :agent_id AND last_event_seq = :expected_last_event_seq" {
		t.Fatalf("projection condition = %q", projectionPut.ConditionExpression)
	}
	assertDynamoDBString(t, projectionPut.Item, "pk", "ASSIGNMENT#asn-000001")
	assertDynamoDBString(t, projectionPut.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBNumber(t, projectionPut.Item, "last_event_seq", "2")
	assertDynamoDBString(t, projectionPut.ExpressionAttributeValues, ":queued", string(AssignmentQueued))
	assertDynamoDBString(t, projectionPut.ExpressionAttributeValues, ":agent_id", "jykim1")
	assertDynamoDBNumber(t, projectionPut.ExpressionAttributeValues, ":expected_last_event_seq", "1")
	if _, ok := projectionPut.Item["agent_id"]; ok {
		t.Fatalf("claimed projection must leave agent_queue: %+v", projectionPut.Item["agent_id"])
	}
	if _, ok := projectionPut.Item["assignment_sort"]; ok {
		t.Fatalf("claimed projection must leave agent_queue: %+v", projectionPut.Item["assignment_sort"])
	}
	activePut := payload.TransactItems[1].Put
	if activePut.ConditionExpression != "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR lease_expires_unix_ms <= :claim_started_unix_ms" {
		t.Fatalf("active condition = %q", activePut.ConditionExpression)
	}
	assertDynamoDBString(t, activePut.Item, "pk", "AGENT#jykim1")
	assertDynamoDBString(t, activePut.Item, "sk", dynamoDBAgentActiveSK)
	assertDynamoDBString(t, activePut.Item, "schema_version", AssignmentAgentActiveSchemaVersion)
	assertDynamoDBString(t, activePut.Item, "agent_id", "jykim1")
	assertDynamoDBString(t, activePut.Item, "active_assignment_id", "asn-000001")
	assertDynamoDBString(t, activePut.Item, "lease_token", wantLeaseToken)
	assertDynamoDBString(t, activePut.Item, "lease_heartbeat_at", claimAt.Format(time.RFC3339Nano))
	assertDynamoDBNumber(t, activePut.Item, "lease_expires_unix_ms", strconv.FormatInt(claimAt.Add(time.Duration(DefaultAssignmentActiveLeaseSeconds)*time.Second).UnixMilli(), 10))
	assertDynamoDBNumber(t, activePut.Item, "last_event_seq", "2")
	assertDynamoDBNumber(t, activePut.ExpressionAttributeValues, ":claim_started_unix_ms", strconv.FormatInt(claimAt.UnixMilli(), 10))
	operationPut := payload.TransactItems[2].Put
	if operationPut.ConditionExpression != "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
		t.Fatalf("operation condition = %q", operationPut.ConditionExpression)
	}
	assertDynamoDBString(t, operationPut.Item, "pk", dynamoDBAssignmentOperationPK)
	assertDynamoDBString(t, operationPut.Item, "operation_type", string(AssignmentOperationPollStart))
	assertDynamoDBString(t, operationPut.Item, "operation_id", claim.Operation.OperationID)
	assertDynamoDBNumber(t, operationPut.Item, "last_event_seq", "2")
}
