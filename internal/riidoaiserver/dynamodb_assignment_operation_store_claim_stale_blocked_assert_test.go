package riidoaiserver

import (
	"strconv"
	"testing"
)

func assertStaleBlockedRepairPayload(
	t *testing.T,
	fixture staleBlockedClaimFixture,
	payload dynamoDBRepairTransactWritePayload,
) {
	t.Helper()
	if len(payload.TransactItems) != 6 {
		t.Fatalf("transact item count = %d", len(payload.TransactItems))
	}
	activeCheck := payload.TransactItems[0].ConditionCheck
	if activeCheck == nil {
		t.Fatalf("missing stale active lease condition: %+v", payload.TransactItems[0])
	}
	assertDynamoDBString(t, activeCheck.Key, "pk", dynamoDBAgentActivePK("jykim-old"))
	assertDynamoDBString(t, activeCheck.ExpressionAttributeValues, ":blocked_assignment_id", fixture.Blocker.ID)
	assertDynamoDBNumber(t, activeCheck.ExpressionAttributeValues, ":claim_started_unix_ms", strconv.FormatInt(fixture.Now.UnixMilli(), 10))
	assertStaleBlockedBlockerRepair(t, fixture, payload)
	assertStaleBlockedCurrentClaim(t, fixture, payload)
}
