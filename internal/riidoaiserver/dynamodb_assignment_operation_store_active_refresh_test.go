package riidoaiserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreRefreshesAgentActiveAssignmentLease(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	refreshAt := fixedNow.Add(10 * time.Second)
	requests := make(chan capturedDynamoDBRequest, 1)
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
		_, _ = w.Write([]byte(`{}`))
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

	assignment := sampleAssignmentOperationRecord(fixedNow).Assignment
	assignment.State = AssignmentRunning
	assignment.LeaseToken = "lease-1"
	if err := store.RefreshAgentActiveAssignment(context.Background(), assignment, refreshAt); err != nil {
		t.Fatalf("RefreshAgentActiveAssignment: %v", err)
	}
	payload := decodeDynamoDBUpdatePayload(t, <-requests)
	if payload.TableName != "assignments" {
		t.Fatalf("update table = %q", payload.TableName)
	}
	if payload.ConditionExpression != "active_assignment_id = :assignment_id AND lease_token = :lease_token" {
		t.Fatalf("condition = %q", payload.ConditionExpression)
	}
	assertDynamoDBString(t, payload.Key, "pk", "AGENT#jykim1")
	assertDynamoDBString(t, payload.Key, "sk", dynamoDBAgentActiveSK)
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":assignment_id", "asn-000001")
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":lease_token", "lease-1")
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":assignment_state", string(AssignmentRunning))
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":heartbeat_at", refreshAt.Format(time.RFC3339Nano))
	assertDynamoDBNumber(t, payload.ExpressionAttributeValues, ":lease_expires_unix_ms", strconv.FormatInt(refreshAt.Add(time.Duration(DefaultAssignmentActiveLeaseSeconds)*time.Second).UnixMilli(), 10))
}
