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

func TestDynamoDBAssignmentOperationStoreSkipsBlockedAssignmentUntilBlockerTerminal(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 3)
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
	activeExpiresAt := fixedNow.Add(time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second)
	activeItem := map[string]map[string]string{
		"schema_version":        {"S": AssignmentAgentActiveSchemaVersion},
		"agent_id":              {"S": "jykim-old"},
		"active_assignment_id":  {"S": "asn-000000"},
		"lease_token":           {"S": "lease-old"},
		"lease_heartbeat_at":    {"S": fixedNow.Format(time.RFC3339Nano)},
		"lease_expires_at":      {"S": activeExpiresAt.Format(time.RFC3339Nano)},
		"lease_expires_unix_ms": {"N": strconv.FormatInt(activeExpiresAt.UnixMilli(), 10)},
	}
	getCalls := 0
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
