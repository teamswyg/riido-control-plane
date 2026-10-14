package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func assertStaleBlockedCurrentClaim(
	t *testing.T,
	fixture staleBlockedClaimFixture,
	payload dynamoDBRepairTransactWritePayload,
) {
	t.Helper()
	currentProjection := payload.TransactItems[3].Put
	if currentProjection == nil {
		t.Fatalf("missing current projection claim put")
	}
	assertDynamoDBString(t, currentProjection.Item, "pk", dynamoDBAssignmentProjectionPK(fixture.Current.ID))
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
	assertDynamoDBString(t, currentActive.Item, "active_assignment_id", fixture.Current.ID)

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
