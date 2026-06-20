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

func TestDynamoDBAssignmentOperationStoreLoadsAgentActiveAssignmentLease(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	expiresAt := fixedNow.Add(time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second)
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
		_ = json.NewEncoder(w).Encode(map[string]any{"Item": map[string]map[string]string{
			"schema_version":        {"S": AssignmentAgentActiveSchemaVersion},
			"agent_id":              {"S": "jykim1"},
			"active_assignment_id":  {"S": "asn-000001"},
			"lease_token":           {"S": "lease-1"},
			"lease_heartbeat_at":    {"S": fixedNow.Format(time.RFC3339Nano)},
			"lease_expires_at":      {"S": expiresAt.Format(time.RFC3339Nano)},
			"lease_expires_unix_ms": {"N": strconv.FormatInt(expiresAt.UnixMilli(), 10)},
		}})
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

	lease, ok, err := store.LoadAgentActiveAssignment(context.Background(), "jykim1")
	if err != nil {
		t.Fatalf("LoadAgentActiveAssignment: %v", err)
	}
	if !ok || lease.ActiveAssignmentID != "asn-000001" || lease.LeaseToken != "lease-1" || lease.Expired(fixedNow) {
		t.Fatalf("lease ok=%v value=%+v", ok, lease)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBGetItemTarget)
}
