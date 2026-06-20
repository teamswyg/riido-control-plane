package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
