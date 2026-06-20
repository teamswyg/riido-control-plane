package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreClaimRepairsStaleBlockedAssignment(t *testing.T) {
	fixedNow := time.Date(2026, 6, 9, 4, 0, 0, 0, time.UTC)
	expiredAt := fixedNow.Add(-time.Minute)
	requests := make(chan capturedDynamoDBRequest, 4)
	current := Assignment{
		ID:                    "asn-000002",
		TaskID:                "task-a",
		ComponentID:           "component-1",
		AgentID:               "jykim-new",
		RuntimeProvider:       "codex",
		Prompt:                "new work",
		State:                 AssignmentQueued,
		ReplacesAssignmentID:  "asn-000001",
		BlockedByAssignmentID: "asn-000001",
		CreatedAt:             fixedNow.Add(-2 * time.Minute),
		UpdatedAt:             fixedNow.Add(-2 * time.Minute),
	}
	blocker := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim-old",
		RuntimeProvider: "codex",
		Prompt:          "old work",
		State:           AssignmentRunning,
		LeaseToken:      "lease-old",
		CreatedAt:       fixedNow.Add(-3 * time.Minute),
		UpdatedAt:       fixedNow.Add(-3 * time.Minute),
	}
	currentItem := sampleAssignmentProjectionDynamoDBItem(t, current, 5)
	blockerItem := sampleAssignmentProjectionDynamoDBItem(t, blocker, 3)
	activeItem := map[string]map[string]string{
		"schema_version":        {"S": AssignmentAgentActiveSchemaVersion},
		"agent_id":              {"S": "jykim-old"},
		"active_assignment_id":  {"S": "asn-000001"},
		"lease_token":           {"S": "lease-old"},
		"lease_heartbeat_at":    {"S": fixedNow.Add(-2 * time.Minute).Format(time.RFC3339Nano)},
		"lease_expires_at":      {"S": expiredAt.Format(time.RFC3339Nano)},
		"lease_expires_unix_ms": {"N": strconv.FormatInt(expiredAt.UnixMilli(), 10)},
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
			if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{currentItem}}); err != nil {
				t.Errorf("encode query response: %v", err)
			}
		case dynamoDBGetItemTarget:
			getCalls++
			item := blockerItem
			if getCalls == 2 {
				item = activeItem
			}
			if err := json.NewEncoder(w).Encode(map[string]any{"Item": item}); err != nil {
				t.Errorf("encode get response: %v", err)
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

	claim, err := store.ClaimNextAssignment(context.Background(), "jykim-new", fixedNow)
	if err != nil {
		t.Fatalf("ClaimNextAssignment repaired: %v", err)
	}
	if !claim.Claimed || claim.Assignment.ID != current.ID || claim.Assignment.State != AssignmentLeased || claim.Assignment.BlockedByAssignmentID != "" {
		t.Fatalf("claim = %+v", claim)
	}
	if len(claim.Operations) != 2 || claim.Operations[0].AssignmentID != blocker.ID || claim.Operations[1].AssignmentID != current.ID {
		t.Fatalf("claim operations = %+v", claim.Operations)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBQueryTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBGetItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBGetItemTarget)
	transactRequest := <-requests
	assertDynamoDBTarget(t, transactRequest, dynamoDBTransactWriteTarget)
	payload := decodeDynamoDBRepairTransactWritePayload(t, transactRequest)
	if len(payload.TransactItems) != 6 {
		t.Fatalf("transact item count = %d", len(payload.TransactItems))
	}
	activeCheck := payload.TransactItems[0].ConditionCheck
	if activeCheck == nil {
		t.Fatalf("missing stale active lease condition: %+v", payload.TransactItems[0])
	}
	assertDynamoDBString(t, activeCheck.Key, "pk", dynamoDBAgentActivePK("jykim-old"))
	assertDynamoDBString(t, activeCheck.ExpressionAttributeValues, ":blocked_assignment_id", blocker.ID)
	assertDynamoDBNumber(t, activeCheck.ExpressionAttributeValues, ":claim_started_unix_ms", strconv.FormatInt(fixedNow.UnixMilli(), 10))

	blockerProjection := payload.TransactItems[1].Put
	if blockerProjection == nil {
		t.Fatalf("missing blocker projection repair put")
	}
	assertDynamoDBString(t, blockerProjection.Item, "pk", dynamoDBAssignmentProjectionPK(blocker.ID))
	assertDynamoDBString(t, blockerProjection.Item, "assignment_state", string(AssignmentFailed))
	assertDynamoDBString(t, blockerProjection.ExpressionAttributeValues, ":expected_state", string(AssignmentRunning))
	assertDynamoDBNumber(t, blockerProjection.ExpressionAttributeValues, ":expected_last_event_seq", "3")

	blockerOperation := payload.TransactItems[2].Put
	if blockerOperation == nil {
		t.Fatalf("missing blocker repair operation put")
	}
	assertDynamoDBString(t, blockerOperation.Item, "operation_type", string(AssignmentOperationAgentEvent))
	assertDynamoDBString(t, blockerOperation.Item, "assignment_state", string(AssignmentFailed))
	var blockerEvents []TaskEvent
	if err := json.Unmarshal([]byte(blockerOperation.Item["events_json"]["S"]), &blockerEvents); err != nil {
		t.Fatalf("decode blocker events: %v", err)
	}
	if len(blockerEvents) != 1 || blockerEvents[0].Type != EventAssignmentFailed || !strings.Contains(blockerEvents[0].Message, "stale blocker") {
		t.Fatalf("blocker events = %+v", blockerEvents)
	}

	currentProjection := payload.TransactItems[3].Put
	if currentProjection == nil {
		t.Fatalf("missing current projection claim put")
	}
	assertDynamoDBString(t, currentProjection.Item, "pk", dynamoDBAssignmentProjectionPK(current.ID))
	assertDynamoDBString(t, currentProjection.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBNumber(t, currentProjection.Item, "last_event_seq", "7")
	if _, ok := currentProjection.Item["blocked_by_assignment_id"]; ok {
		t.Fatalf("claimed projection must clear blocker: %+v", currentProjection.Item)
	}
	if _, ok := currentProjection.Item["agent_id"]; ok {
		t.Fatalf("claimed projection must leave agent_queue: %+v", currentProjection.Item)
	}

	currentActive := payload.TransactItems[4].Put
	if currentActive == nil {
		t.Fatalf("missing current active lease put")
	}
	assertDynamoDBString(t, currentActive.Item, "pk", dynamoDBAgentActivePK("jykim-new"))
	assertDynamoDBString(t, currentActive.Item, "active_assignment_id", current.ID)

	currentOperation := payload.TransactItems[5].Put
	if currentOperation == nil {
		t.Fatalf("missing current claim operation put")
	}
	assertDynamoDBString(t, currentOperation.Item, "operation_type", string(AssignmentOperationPollStart))
	assertDynamoDBNumber(t, currentOperation.Item, "last_event_seq", "7")
	var currentEvents []TaskEvent
	if err := json.Unmarshal([]byte(currentOperation.Item["events_json"]["S"]), &currentEvents); err != nil {
		t.Fatalf("decode current events: %v", err)
	}
	if len(currentEvents) != 2 || currentEvents[0].Type != EventAssignmentQueued || currentEvents[1].Type != EventAssignmentLeased {
		t.Fatalf("current events = %+v", currentEvents)
	}
}
