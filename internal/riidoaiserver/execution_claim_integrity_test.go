package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestExecutionClaimRejectsReplacementAndMalformedJournal(t *testing.T) {
	now := time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC)
	a := sampleQueuedAssignmentOperationRecord(now).Assignment
	record := dynamoDBClaimAssignmentOperation(a, 1, now, "", nil)
	ctx := context.WithValue(context.Background(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{"device", "owner"})
	record.ExecutionClaim = newExecutionClaim(ctx, record.Assignment)
	state, err := stateFromAssignmentOperations([]AssignmentOperationRecord{record})
	if err != nil {
		t.Fatal(err)
	}
	changed := *record.ExecutionClaim
	changed.OwnerPrincipalID = "replacement"
	if err := rememberExecutionClaim(&state, &changed); err == nil {
		t.Fatal("claim owner replaced")
	}
	item, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := assignmentOperationRecordFromDynamoDBItem(item)
	if err != nil || decoded.ExecutionClaim == nil || *decoded.ExecutionClaim != *record.ExecutionClaim {
		t.Fatal("journal item lost security claim")
	}
	decoded.ExecutionClaim.RequesterPrincipalID = "forged"
	if _, err := stateFromAssignmentOperations([]AssignmentOperationRecord{decoded}); err == nil {
		t.Fatal("journal requester mismatch accepted")
	}
}

func TestExecutionClaimTerminalBudgetRejectsExtraOrUnboundedActions(t *testing.T) {
	claim := executionClaim{DeviceID: "device", OwnerPrincipalID: "owner", AssignmentID: "assignment", AgentID: "agent", LeaseEpoch: leaseEpoch("fixture"), ClaimedAt: time.Now()}
	item, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	raw := func(item map[string]map[string]string) json.RawMessage {
		b, e := json.Marshal(dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{Item: item, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}})
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	sidecar := raw(item)
	if err := validateExecutionClaimTerminalBudget([]json.RawMessage{sidecar, sidecar}); err == nil {
		t.Fatal("two sidecars accepted")
	}
	item["extra"] = map[string]string{"S": strings.Repeat("x", executionClaimItemBytes)}
	if err := validateExecutionClaimTerminalBudget([]json.RawMessage{raw(item)}); err == nil {
		t.Fatal("unbounded sidecar accepted")
	}
	arbitrary := make([]json.RawMessage, 7)
	for i := range arbitrary {
		arbitrary[i] = raw(map[string]map[string]string{"pk": {"S": "arbitrary"}, "sk": {"S": "fixture"}})
	}
	if err := validateExecutionClaimTerminalBudget(arbitrary); err == nil {
		t.Fatal("arbitrary action allowance increased")
	}
}
