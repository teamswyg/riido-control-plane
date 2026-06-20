package riidoaiserver

import (
	"strconv"
	"testing"
	"time"
)

func assertQueuedClaimTransactPayload(
	t *testing.T,
	claim AssignmentClaimResult,
	claimAt time.Time,
	payload dynamoDBTransactWritePayload,
) {
	t.Helper()
	if len(payload.TransactItems) != 3 {
		t.Fatalf("transact item count = %d", len(payload.TransactItems))
	}
	wantLeaseToken := "asn-000001:" + strconv.FormatInt(claimAt.UnixNano(), 10)
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
	assertQueuedClaimActivePut(t, claimAt, wantLeaseToken, payload.TransactItems[1].Put)
	operationPut := payload.TransactItems[2].Put
	if operationPut.ConditionExpression != "attribute_not_exists(pk) AND attribute_not_exists(sk)" {
		t.Fatalf("operation condition = %q", operationPut.ConditionExpression)
	}
	assertDynamoDBString(t, operationPut.Item, "pk", dynamoDBAssignmentOperationPK)
	assertDynamoDBString(t, operationPut.Item, "operation_type", string(AssignmentOperationPollStart))
	assertDynamoDBString(t, operationPut.Item, "operation_id", claim.Operation.OperationID)
	assertDynamoDBNumber(t, operationPut.Item, "last_event_seq", "2")
}
