package riidoaiserver

import (
	"encoding/json"
	"testing"
)

type dynamodbPutItemPayload struct {
	TableName           string                       `json:"TableName"`
	ConditionExpression string                       `json:"ConditionExpression"`
	Item                map[string]map[string]string `json:"Item"`
}

func decodeDynamoDBOutboxPutPayload(t *testing.T, body []byte) dynamodbPutItemPayload {
	t.Helper()
	var payload dynamodbPutItemPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}

func assertDynamoDBOutboxPutPayload(t *testing.T, payload dynamodbPutItemPayload) {
	t.Helper()
	if payload.TableName != "riido-ai-server-event-outbox" {
		t.Fatalf("table = %q", payload.TableName)
	}
	if payload.ConditionExpression != "attribute_not_exists(task_id) AND attribute_not_exists(event_seq)" {
		t.Fatalf("condition = %q", payload.ConditionExpression)
	}
	assertDynamoDBString(t, payload.Item, "task_id", "task-a")
	assertDynamoDBNumber(t, payload.Item, "event_seq", "7")
	assertDynamoDBString(t, payload.Item, "assignment_id", "assignment-1")
	assertDynamoDBString(t, payload.Item, "agent_id", "jykim1")
	assertDynamoDBString(t, payload.Item, "event_type", EventAssignmentLeased)
	assertDynamoDBString(t, payload.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBString(t, payload.Item, "message", "leased")
	assertDynamoDBString(t, payload.Item, "metadata_json", `{"lease_token":"lease-1"}`)
	assertDynamoDBString(t, payload.Item, "schema_version", OutboxRecordSchemaVersion)
	assertDynamoDBString(t, payload.Item, "at", "2026-05-26T01:02:03Z")
	assertDynamoDBOutboxRecord(t, payload.Item["event_json"]["S"])
}

func assertDynamoDBOutboxRecord(t *testing.T, raw string) {
	t.Helper()
	var record OutboxRecord
	if err := json.Unmarshal([]byte(raw), &record); err != nil {
		t.Fatalf("decode event_json: %v", err)
	}
	if record.SchemaVersion != OutboxRecordSchemaVersion ||
		record.Event.TaskID != "task-a" ||
		record.Event.Type != EventAssignmentLeased {
		t.Fatalf("record = %+v", record)
	}
}
