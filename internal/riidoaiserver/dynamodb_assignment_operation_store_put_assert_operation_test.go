package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func assertAssignmentOperationJournalPutPayload(
	t *testing.T,
	payload dynamoDBPutPayload,
) {
	t.Helper()
	if payload.TableName != "riido-ai-server-assignments" {
		t.Fatalf("operation table = %q", payload.TableName)
	}
	if payload.ConditionExpression != "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
		t.Fatalf("operation condition = %q", payload.ConditionExpression)
	}
	assertDynamoDBString(t, payload.Item, "pk", dynamoDBAssignmentOperationPK)
	assertDynamoDBString(t, payload.Item, "sk", "20260526T010203.000000000Z#00000000000000000002#poll-start:asn-000001:lease-1:2")
	assertDynamoDBString(t, payload.Item, "schema_version", AssignmentOperationSchemaVersion)
	assertDynamoDBString(t, payload.Item, "operation_id", "poll-start:asn-000001:lease-1:2")
	assertDynamoDBString(t, payload.Item, "operation_type", string(AssignmentOperationPollStart))
	assertDynamoDBString(t, payload.Item, "task_id", "task-a")
	assertDynamoDBString(t, payload.Item, "assignment_id", "asn-000001")
	assertDynamoDBString(t, payload.Item, "operation_agent_id", "jykim1")
	assertDynamoDBString(t, payload.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBString(t, payload.Item, "recorded_at", "2026-05-26T01:02:03Z")
	assertDynamoDBNumber(t, payload.Item, "last_event_seq", "2")
	assertDynamoDBNumber(t, payload.Item, "event_count", "1")
	if _, ok := payload.Item["agent_id"]; ok {
		t.Fatalf("operation journal item must not project into agent_queue GSI: %+v", payload.Item["agent_id"])
	}
	if _, ok := payload.Item["assignment_sort"]; ok {
		t.Fatalf("operation journal item must not project into agent_queue GSI: %+v", payload.Item["assignment_sort"])
	}
	assertAssignmentOperationJournalJSON(t, payload)
}

func assertAssignmentOperationJournalJSON(t *testing.T, payload dynamoDBPutPayload) {
	t.Helper()
	var assignment Assignment
	if err := json.Unmarshal([]byte(payload.Item["assignment_json"]["S"]), &assignment); err != nil {
		t.Fatalf("decode assignment_json: %v", err)
	}
	if assignment.ID != "asn-000001" || assignment.LeaseToken != "lease-1" {
		t.Fatalf("assignment_json = %+v", assignment)
	}
	var events []TaskEvent
	if err := json.Unmarshal([]byte(payload.Item["events_json"]["S"]), &events); err != nil {
		t.Fatalf("decode events_json: %v", err)
	}
	if len(events) != 1 || events[0].Seq != 2 || events[0].Type != EventAssignmentLeased {
		t.Fatalf("events_json = %+v", events)
	}
}
