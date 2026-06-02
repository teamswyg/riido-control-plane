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

func TestDynamoDBAssignmentOperationStoreWritesPutItem(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 2)
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

	provider, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "SESSION")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-server-assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAssignmentOperationStore: %v", err)
	}
	defer store.Close()

	record := sampleAssignmentOperationRecord(fixedNow)
	if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatalf("SaveAssignmentOperation: %v", err)
	}

	first := <-requests
	second := <-requests
	operationPayload := decodeDynamoDBPutPayload(t, first)
	projectionPayload := decodeDynamoDBPutPayload(t, second)
	if operationPayload.Item["pk"]["S"] != dynamoDBAssignmentOperationPK {
		operationPayload, projectionPayload = projectionPayload, operationPayload
	}

	if operationPayload.TableName != "riido-ai-server-assignments" {
		t.Fatalf("operation table = %q", operationPayload.TableName)
	}
	if operationPayload.ConditionExpression != "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
		t.Fatalf("operation condition = %q", operationPayload.ConditionExpression)
	}
	assertDynamoDBString(t, operationPayload.Item, "pk", dynamoDBAssignmentOperationPK)
	assertDynamoDBString(t, operationPayload.Item, "sk", "20260526T010203.000000000Z#00000000000000000002#poll-start:asn-000001:lease-1:2")
	assertDynamoDBString(t, operationPayload.Item, "schema_version", AssignmentOperationSchemaVersion)
	assertDynamoDBString(t, operationPayload.Item, "operation_id", "poll-start:asn-000001:lease-1:2")
	assertDynamoDBString(t, operationPayload.Item, "operation_type", string(AssignmentOperationPollStart))
	assertDynamoDBString(t, operationPayload.Item, "task_id", "task-a")
	assertDynamoDBString(t, operationPayload.Item, "assignment_id", "asn-000001")
	assertDynamoDBString(t, operationPayload.Item, "operation_agent_id", "jykim1")
	assertDynamoDBString(t, operationPayload.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBString(t, operationPayload.Item, "recorded_at", "2026-05-26T01:02:03Z")
	assertDynamoDBNumber(t, operationPayload.Item, "last_event_seq", "2")
	assertDynamoDBNumber(t, operationPayload.Item, "event_count", "1")
	if _, ok := operationPayload.Item["agent_id"]; ok {
		t.Fatalf("operation journal item must not project into agent_queue GSI: %+v", operationPayload.Item["agent_id"])
	}
	if _, ok := operationPayload.Item["assignment_sort"]; ok {
		t.Fatalf("operation journal item must not project into agent_queue GSI: %+v", operationPayload.Item["assignment_sort"])
	}

	var assignment Assignment
	if err := json.Unmarshal([]byte(operationPayload.Item["assignment_json"]["S"]), &assignment); err != nil {
		t.Fatalf("decode assignment_json: %v", err)
	}
	if assignment.ID != "asn-000001" || assignment.LeaseToken != "lease-1" {
		t.Fatalf("assignment_json = %+v", assignment)
	}
	var events []TaskEvent
	if err := json.Unmarshal([]byte(operationPayload.Item["events_json"]["S"]), &events); err != nil {
		t.Fatalf("decode events_json: %v", err)
	}
	if len(events) != 1 || events[0].Seq != 2 || events[0].Type != EventAssignmentLeased {
		t.Fatalf("events_json = %+v", events)
	}

	if projectionPayload.TableName != "riido-ai-server-assignments" {
		t.Fatalf("projection table = %q", projectionPayload.TableName)
	}
	if projectionPayload.ConditionExpression != "attribute_not_exists(last_event_seq) OR last_event_seq <= :last_event_seq" {
		t.Fatalf("projection condition = %q", projectionPayload.ConditionExpression)
	}
	assertDynamoDBString(t, projectionPayload.Item, "pk", "ASSIGNMENT#asn-000001")
	assertDynamoDBString(t, projectionPayload.Item, "sk", dynamoDBAssignmentProjectionSK)
	assertDynamoDBString(t, projectionPayload.Item, "schema_version", AssignmentProjectionSchemaVersion)
	assertDynamoDBString(t, projectionPayload.Item, "assignment_id", "asn-000001")
	assertDynamoDBString(t, projectionPayload.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBNumber(t, projectionPayload.Item, "last_event_seq", "2")
	assertDynamoDBNumber(t, projectionPayload.ExpressionAttributeValues, ":last_event_seq", "2")
	if _, ok := projectionPayload.Item["agent_id"]; ok {
		t.Fatalf("leased assignment projection must not remain in agent queue: %+v", projectionPayload.Item["agent_id"])
	}
	if _, ok := projectionPayload.Item["assignment_sort"]; ok {
		t.Fatalf("leased assignment projection must not remain in agent queue: %+v", projectionPayload.Item["assignment_sort"])
	}
}

func TestDynamoDBAssignmentOperationStoreQueriesReplayableJournal(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	queryRequests := make(chan []byte, 1)
	record := sampleAssignmentOperationRecord(fixedNow)
	item := sampleAssignmentOperationDynamoDBItem(t, record)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if r.Header.Get("X-Amz-Target") != dynamoDBQueryTarget {
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
		}
		queryRequests <- body
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{item}}); err != nil {
			t.Errorf("encode response: %v", err)
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

	records, err := store.LoadAssignmentOperations(context.Background())
	if err != nil {
		t.Fatalf("LoadAssignmentOperations: %v", err)
	}
	if len(records) != 1 || records[0].OperationID != record.OperationID || records[0].Assignment.ID != record.Assignment.ID {
		t.Fatalf("records = %+v", records)
	}
	var payload struct {
		TableName                 string                       `json:"TableName"`
		ConsistentRead            bool                         `json:"ConsistentRead"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}
	if err := json.Unmarshal(<-queryRequests, &payload); err != nil {
		t.Fatalf("decode query payload: %v", err)
	}
	if payload.TableName != "assignments" || !payload.ConsistentRead || !payload.ScanIndexForward || payload.Limit != dynamoDBAssignmentQueryLimit {
		t.Fatalf("query payload = %+v", payload)
	}
	if payload.KeyConditionExpression != "pk = :pk" {
		t.Fatalf("key condition = %q", payload.KeyConditionExpression)
	}
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":pk", dynamoDBAssignmentOperationPK)
}

func TestDynamoDBAssignmentOperationStoreProjectsQueuedAssignmentIntoAgentQueue(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 2)
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
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAssignmentOperationStore: %v", err)
	}
	defer store.Close()

	record := sampleQueuedAssignmentOperationRecord(fixedNow)
	if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatalf("SaveAssignmentOperation: %v", err)
	}
	first := decodeDynamoDBPutPayload(t, <-requests)
	second := decodeDynamoDBPutPayload(t, <-requests)
	projection := second
	if first.Item["pk"]["S"] != dynamoDBAssignmentOperationPK {
		projection = first
	}
	assertDynamoDBString(t, projection.Item, "pk", "ASSIGNMENT#asn-000001")
	assertDynamoDBString(t, projection.Item, "sk", dynamoDBAssignmentProjectionSK)
	assertDynamoDBString(t, projection.Item, "agent_id", "jykim1")
	assertDynamoDBString(t, projection.Item, "assignment_sort", "20260526T010203.000000000Z#asn-000001")
}

func TestDynamoDBAssignmentOperationStoreQueriesAgentQueue(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	queryRequests := make(chan []byte, 1)
	assignment := sampleQueuedAssignmentOperationRecord(fixedNow).Assignment
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if r.Header.Get("X-Amz-Target") != dynamoDBQueryTarget {
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
		}
		queryRequests <- body
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		if err := json.NewEncoder(w).Encode(map[string]any{"Items": []map[string]map[string]string{item}}); err != nil {
			t.Errorf("encode response: %v", err)
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

	assignments, err := store.LoadAgentQueueAssignments(context.Background(), "jykim1")
	if err != nil {
		t.Fatalf("LoadAgentQueueAssignments: %v", err)
	}
	if len(assignments) != 1 || assignments[0].ID != assignment.ID || assignments[0].State != AssignmentQueued {
		t.Fatalf("assignments = %+v", assignments)
	}
	var payload struct {
		TableName                 string                       `json:"TableName"`
		IndexName                 string                       `json:"IndexName"`
		KeyConditionExpression    string                       `json:"KeyConditionExpression"`
		ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
		ScanIndexForward          bool                         `json:"ScanIndexForward"`
		Limit                     int                          `json:"Limit"`
	}
	if err := json.Unmarshal(<-queryRequests, &payload); err != nil {
		t.Fatalf("decode query payload: %v", err)
	}
	if payload.TableName != "assignments" || payload.IndexName != dynamoDBAssignmentQueueIndex || !payload.ScanIndexForward || payload.Limit != dynamoDBAssignmentQueryLimit {
		t.Fatalf("query payload = %+v", payload)
	}
	if payload.KeyConditionExpression != "agent_id = :agent_id" {
		t.Fatalf("key condition = %q", payload.KeyConditionExpression)
	}
	assertDynamoDBString(t, payload.ExpressionAttributeValues, ":agent_id", "jykim1")
}

func TestDynamoDBAssignmentOperationStoreLoadsAssignmentProjection(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 1)
	assignment := sampleAssignmentOperationRecord(fixedNow).Assignment
	assignment.State = AssignmentCancelling
	item := sampleAssignmentProjectionDynamoDBItem(t, assignment, 7)
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
		_ = json.NewEncoder(w).Encode(map[string]any{"Item": item})
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

	projection, ok, err := store.LoadAssignmentProjection(context.Background(), assignment.ID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection: %v", err)
	}
	if !ok || projection.Assignment.ID != assignment.ID || projection.Assignment.State != AssignmentCancelling || projection.LastEventSeq != 7 {
		t.Fatalf("projection ok=%v value=%+v", ok, projection)
	}
	got := <-requests
	assertDynamoDBTarget(t, got, dynamoDBGetItemTarget)
	var payload struct {
		TableName string                       `json:"TableName"`
		Key       map[string]map[string]string `json:"Key"`
	}
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode get payload: %v", err)
	}
	assertDynamoDBString(t, payload.Key, "pk", "ASSIGNMENT#asn-000001")
	assertDynamoDBString(t, payload.Key, "sk", dynamoDBAssignmentProjectionSK)
}

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

func TestDynamoDBAssignmentOperationStoreClaimContentionReturnsNoClaim(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
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
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#TransactionCanceledException","message":"condition failed"}`))
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
		t.Fatalf("ClaimNextAssignment contention: %v", err)
	}
	if claim.Claimed {
		t.Fatalf("claim should be empty after contention: %+v", claim)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBQueryTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBTransactWriteTarget)
}

func TestDynamoDBAssignmentOperationStoreSkipsBlockedAssignmentUntilBlockerTerminal(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 2)
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
			if err := json.NewEncoder(w).Encode(map[string]any{"Item": blockerItem}); err != nil {
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
}

func TestDynamoDBAssignmentOperationStoreReleasesAgentActiveAssignmentOnTerminalEvent(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 3)
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

	record := sampleTerminalAssignmentOperationRecord(fixedNow)
	if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatalf("SaveAssignmentOperation terminal: %v", err)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	deletePayload := decodeDynamoDBDeletePayload(t, <-requests)
	if deletePayload.TableName != "assignments" {
		t.Fatalf("delete table = %q", deletePayload.TableName)
	}
	if deletePayload.ConditionExpression != "active_assignment_id = :assignment_id" {
		t.Fatalf("delete condition = %q", deletePayload.ConditionExpression)
	}
	assertDynamoDBString(t, deletePayload.Key, "pk", "AGENT#jykim1")
	assertDynamoDBString(t, deletePayload.Key, "sk", dynamoDBAgentActiveSK)
	assertDynamoDBString(t, deletePayload.ExpressionAttributeValues, ":assignment_id", "asn-000001")
}

func TestDynamoDBAssignmentOperationStoreTreatsMissingAgentActiveReleaseAsIdempotent(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	requests := make(chan capturedDynamoDBRequest, 3)
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
		if r.Header.Get("X-Amz-Target") == dynamoDBDeleteItemTarget {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","message":"missing active row"}`))
			return
		}
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

	if err := store.SaveAssignmentOperation(context.Background(), sampleTerminalAssignmentOperationRecord(fixedNow)); err != nil {
		t.Fatalf("SaveAssignmentOperation terminal with missing active row: %v", err)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBDeleteItemTarget)
}

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

func TestDynamoDBAssignmentOperationStoreTreatsConditionalCheckFailedAsIdempotent(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		var payload struct {
			Item map[string]map[string]string `json:"Item"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode request: %v", err)
		}
		pk := payload.Item["pk"]["S"]
		requests <- pk
		if pk == dynamoDBAssignmentOperationPK {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","message":"duplicate"}`))
			return
		}
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
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAssignmentOperationStore: %v", err)
	}
	defer store.Close()

	err = store.SaveAssignmentOperation(context.Background(), sampleAssignmentOperationRecord(time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)))
	if err != nil {
		t.Fatalf("SaveAssignmentOperation duplicate: %v", err)
	}
	if first, second := <-requests, <-requests; first != dynamoDBAssignmentOperationPK || second != "ASSIGNMENT#asn-000001" {
		t.Fatalf("request pks = %q, %q", first, second)
	}
}

func sampleAssignmentOperationDynamoDBItem(t *testing.T, record AssignmentOperationRecord) map[string]map[string]string {
	t.Helper()
	assignmentJSON, err := json.Marshal(record.Assignment)
	if err != nil {
		t.Fatalf("marshal assignment: %v", err)
	}
	eventsJSON, err := json.Marshal(record.Events)
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	return map[string]map[string]string{
		"pk":                 {"S": dynamoDBAssignmentOperationPK},
		"sk":                 {"S": assignmentOperationSortKey(record)},
		"schema_version":     {"S": record.SchemaVersion},
		"operation_id":       {"S": record.OperationID},
		"operation_type":     {"S": string(record.OperationType)},
		"task_id":            {"S": record.TaskID},
		"assignment_id":      {"S": record.AssignmentID},
		"operation_agent_id": {"S": record.AgentID},
		"recorded_at":        {"S": record.RecordedAt.UTC().Format(time.RFC3339Nano)},
		"assignment_json":    {"S": string(assignmentJSON)},
		"events_json":        {"S": string(eventsJSON)},
	}
}

func sampleAssignmentProjectionDynamoDBItem(t *testing.T, assignment Assignment, lastEventSeq int64) map[string]map[string]string {
	t.Helper()
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		t.Fatalf("marshal assignment: %v", err)
	}
	return map[string]map[string]string{
		"pk":              {"S": dynamoDBAssignmentProjectionPK(assignment.ID)},
		"sk":              {"S": dynamoDBAssignmentProjectionSK},
		"schema_version":  {"S": AssignmentProjectionSchemaVersion},
		"assignment_id":   {"S": assignment.ID},
		"agent_id":        {"S": assignment.AgentID},
		"assignment_sort": {"S": assignmentQueueSort(assignment)},
		"last_event_seq":  {"N": strconv.FormatInt(lastEventSeq, 10)},
		"assignment_json": {"S": string(assignmentJSON)},
	}
}

func TestDynamoDBAssignmentOperationStoreRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "ap-northeast-2", TableName: "assignments"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
}

func sampleAssignmentOperationRecord(at time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
		State:           AssignmentLeased,
		LeaseToken:      "lease-1",
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	events := []TaskEvent{{
		Seq:          2,
		TaskID:       "task-a",
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentLeased,
		State:        AssignmentLeased,
		Metadata:     map[string]string{"lease_token": assignment.LeaseToken},
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationPollStart, assignment, events),
		OperationType: AssignmentOperationPollStart,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}

func sampleTerminalAssignmentOperationRecord(at time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
		State:           AssignmentCompleted,
		LeaseToken:      "lease-1",
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	events := []TaskEvent{{
		Seq:          3,
		TaskID:       "task-a",
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentCompleted,
		State:        AssignmentCompleted,
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAgentEvent, assignment, events),
		OperationType: AssignmentOperationAgentEvent,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}

func sampleQueuedAssignmentOperationRecord(at time.Time) AssignmentOperationRecord {
	assignment := Assignment{
		ID:              "asn-000001",
		TaskID:          "task-a",
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "make hello world",
		State:           AssignmentQueued,
		CreatedAt:       at,
		UpdatedAt:       at,
	}
	events := []TaskEvent{{
		Seq:          1,
		TaskID:       "task-a",
		AssignmentID: assignment.ID,
		AgentID:      assignment.AgentID,
		Type:         EventAssignmentQueued,
		State:        AssignmentQueued,
		At:           at,
	}}
	return AssignmentOperationRecord{
		SchemaVersion: AssignmentOperationSchemaVersion,
		OperationID:   assignmentOperationID(AssignmentOperationAssignTask, assignment, events),
		OperationType: AssignmentOperationAssignTask,
		TaskID:        assignment.TaskID,
		AssignmentID:  assignment.ID,
		AgentID:       assignment.AgentID,
		Assignment:    assignment,
		Events:        events,
		RecordedAt:    at,
	}
}

type dynamoDBPutPayload struct {
	TableName                 string                       `json:"TableName"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	Item                      map[string]map[string]string `json:"Item"`
}

type dynamoDBTransactWritePayload struct {
	TransactItems []struct {
		Put dynamoDBTransactPut `json:"Put"`
	} `json:"TransactItems"`
}

type dynamoDBTransactPut struct {
	TableName                 string                       `json:"TableName"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
	Item                      map[string]map[string]string `json:"Item"`
}

type dynamoDBDeletePayload struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
}

type dynamoDBUpdatePayload struct {
	TableName                 string                       `json:"TableName"`
	Key                       map[string]map[string]string `json:"Key"`
	ConditionExpression       string                       `json:"ConditionExpression"`
	UpdateExpression          string                       `json:"UpdateExpression"`
	ExpressionAttributeValues map[string]map[string]string `json:"ExpressionAttributeValues"`
}

func decodeDynamoDBPutPayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBPutPayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBPutItemTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBPutPayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}

func decodeDynamoDBDeletePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBDeletePayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBDeleteItemTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBDeletePayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode delete payload: %v", err)
	}
	return payload
}

func decodeDynamoDBUpdatePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBUpdatePayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBUpdateItemTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBUpdatePayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode update payload: %v", err)
	}
	return payload
}

func decodeDynamoDBTransactWritePayload(t *testing.T, got capturedDynamoDBRequest) dynamoDBTransactWritePayload {
	t.Helper()
	if got.method != http.MethodPost {
		t.Fatalf("method = %s", got.method)
	}
	if got.header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
		t.Fatalf("target = %q", got.header.Get("X-Amz-Target"))
	}
	var payload dynamoDBTransactWritePayload
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode transact payload: %v", err)
	}
	return payload
}
