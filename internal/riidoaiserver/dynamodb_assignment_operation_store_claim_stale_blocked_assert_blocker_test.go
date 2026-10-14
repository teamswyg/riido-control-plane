package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func assertStaleBlockedBlockerRepair(
	t *testing.T,
	fixture staleBlockedClaimFixture,
	payload dynamoDBRepairTransactWritePayload,
) {
	t.Helper()
	blockerProjection := payload.TransactItems[1].Put
	if blockerProjection == nil {
		t.Fatalf("missing blocker projection repair put")
	}
	assertDynamoDBString(t, blockerProjection.Item, "pk", dynamoDBAssignmentProjectionPK(fixture.Blocker.ID))
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
}
