package riidoaiserver

import "testing"

func assertAssignmentProjectionPutPayload(t *testing.T, payload dynamoDBPutPayload) {
	t.Helper()
	if payload.TableName != "riido-ai-server-assignments" {
		t.Fatalf("projection table = %q", payload.TableName)
	}
	if payload.ConditionExpression != "attribute_not_exists(last_event_seq) OR last_event_seq <= :last_event_seq" {
		t.Fatalf("projection condition = %q", payload.ConditionExpression)
	}
	assertDynamoDBString(t, payload.Item, "pk", "ASSIGNMENT#asn-000001")
	assertDynamoDBString(t, payload.Item, "sk", dynamoDBAssignmentProjectionSK)
	assertDynamoDBString(t, payload.Item, "schema_version", AssignmentProjectionSchemaVersion)
	assertDynamoDBString(t, payload.Item, "assignment_id", "asn-000001")
	assertDynamoDBString(t, payload.Item, "assignment_state", string(AssignmentLeased))
	assertDynamoDBNumber(t, payload.Item, "last_event_seq", "2")
	assertDynamoDBNumber(t, payload.ExpressionAttributeValues, ":last_event_seq", "2")
	if _, ok := payload.Item["agent_id"]; ok {
		t.Fatalf("leased assignment projection must not remain in agent queue: %+v", payload.Item["agent_id"])
	}
	if _, ok := payload.Item["assignment_sort"]; ok {
		t.Fatalf("leased assignment projection must not remain in agent queue: %+v", payload.Item["assignment_sort"])
	}
}
