package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestIdentifiedAssignmentTransactionCannotOverwriteWinner(t *testing.T) {
	var mu sync.Mutex
	var projection map[string]map[string]string
	var writes int
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget {
			json.NewEncoder(w).Encode(map[string]any{"Item": projection})
			return
		}
		if r.Header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
			t.Error("creation must be atomic")
			http.Error(w, "bad target", 400)
			return
		}
		var payload struct {
			TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if len(payload.TransactItems) != 2 {
			t.Error("operation and projection must share transaction")
			return
		}
		for _, item := range payload.TransactItems {
			if item.Put.ConditionExpression != "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
				t.Error("missing create-only guard")
			}
		}
		if writes > 0 {
			w.WriteHeader(400)
			w.Write([]byte(`{"__type":"TransactionCanceledException"}`))
			return
		}
		writes++
		projection = payload.TransactItems[0].Put.Item
		// Simulate a daemon completing the winner before the second replica retries.
		var assignment Assignment
		json.Unmarshal([]byte(projection["assignment_json"]["S"]), &assignment)
		assignment.State = AssignmentCompleted
		value, _ := json.Marshal(assignment)
		projection["assignment_json"] = map[string]string{"S": string(value)}
		projection["assignment_state"] = map[string]string{"S": string(AssignmentCompleted)}
		w.Write([]byte(`{}`))
	}))
	defer endpoint.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test-region", TableName: "test-assignments", Endpoint: endpoint.URL, CredentialsProvider: credentials})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	record := sampleAssignmentOperationRecord(time.Now())
	record.OperationType = AssignmentOperationAssignTask
	record.OperationID = "assign:request"
	record.Assignment.RequestID = "request"
	record.Assignment.State = AssignmentQueued
	for range 2 {
		if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
			t.Fatal(err)
		}
	}
	result, found, err := store.LoadAssignmentProjection(context.Background(), record.AssignmentID)
	if err != nil || !found || result.Assignment.State != AssignmentCompleted || writes != 1 {
		t.Fatalf("winner overwritten: %+v %v writes=%d", result, err, writes)
	}
}

func TestClientSnapshotCompareAndSwapRejectsStaleWriter(t *testing.T) {
	var version int64
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct{ TransactItems []dynamoDBTransactWritePut }
		json.NewDecoder(r.Body).Decode(&payload)
		if len(payload.TransactItems) == 0 {
			t.Error("missing transaction")
			return
		}
		p := payload.TransactItems[len(payload.TransactItems)-1].Put
		if p.ConditionExpression == "" {
			t.Error("missing version check")
		}
		if p.ExpressionAttributeValues[":expected"]["N"] != "0" {
			t.Error("unexpected previous version")
		}
		if version > 0 {
			w.WriteHeader(400)
			w.Write([]byte(`{"__type":"ConditionalCheckFailedException"}`))
			return
		}
		version++
		w.Write([]byte(`{}`))
	}))
	defer endpoint.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{Region: "test-region", TableName: "test-client", Endpoint: endpoint.URL, CredentialsProvider: credentials})
	if err != nil {
		t.Fatal(err)
	}
	snapshot := AIAgentClientSnapshot{SchemaVersion: AIAgentClientPersistenceSchemaVersion, Version: 1, ExpectedVersion: 0}
	if err := store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAIAgentClientSnapshot(context.Background(), snapshot); err == nil {
		t.Fatal("stale writer silently overwrote membership")
	}
}

func TestDurableCancellationTransactionUsesExpectedStateSequenceAndLease(t *testing.T) {
	calls := 0
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
			t.Error("cancellation log and projection must be atomic")
		}
		var payload struct{ TransactItems []dynamoDBTransactWritePut }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.TransactItems) != 2 {
			t.Errorf("transaction items=%d", len(payload.TransactItems))
			return
		}
		guard := payload.TransactItems[0].Put
		if guard.ConditionExpression != "assignment_state = :expected_state AND last_event_seq = :expected_seq AND lease_token = :expected_lease" || guard.ExpressionAttributeValues[":expected_seq"]["N"] != "5" || guard.ExpressionAttributeValues[":expected_state"]["S"] != string(AssignmentRunning) || guard.ExpressionAttributeValues[":expected_lease"]["S"] != "lease" {
			t.Errorf("incomplete guard: %+v", guard)
		}
		if calls == 2 {
			w.WriteHeader(400)
			w.Write([]byte(`{"__type":"TransactionCanceledException"}`))
			return
		}
		w.Write([]byte(`{}`))
	}))
	defer endpoint.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	record := sampleAssignmentOperationRecord(time.Now())
	record.OperationType = AssignmentOperationClientStop
	record.Assignment.State = AssignmentCancelling
	record.Assignment.LeaseToken = "lease"
	expected := record.Assignment
	expected.State = AssignmentRunning
	record.ExpectedProjection = &AssignmentProjection{Assignment: expected, LastEventSeq: 5}
	if err = store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	if err = store.SaveAssignmentOperation(context.Background(), record); !errors.Is(err, errAssignmentProjectionConflict) {
		t.Fatalf("conflict must retry: %v", err)
	}
}
