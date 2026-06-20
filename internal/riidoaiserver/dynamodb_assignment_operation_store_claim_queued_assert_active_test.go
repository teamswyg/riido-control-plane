package riidoaiserver

import (
	"strconv"
	"testing"
	"time"
)

func assertQueuedClaimActivePut(
	t *testing.T,
	claimAt time.Time,
	wantLeaseToken string,
	activePut dynamoDBTransactPut,
) {
	t.Helper()
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
}
