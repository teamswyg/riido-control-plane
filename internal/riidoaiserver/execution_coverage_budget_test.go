package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestCoverageReadyFencePreservesBoundedTransaction(t *testing.T) {
	coverage, err := executionCoverageFromItem(coverageRevisionFixture(), "dcov_fixture")
	if err != nil {
		t.Fatal(err)
	}
	action := dynamoDBTransactWriteItem{ConditionCheck: executionCoverageCondition("fixture", coverage)}
	raw, err := json.Marshal(action)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateExecutionClaimTerminalBudget([]json.RawMessage{raw}); err != nil {
		t.Fatalf("valid Ready authority condition cannot reach persistence: %v", err)
	}
	action.ConditionCheck.ExpressionAttributeValues[":revision"] = map[string]string{"N": "0"}
	raw, _ = json.Marshal(action)
	if err := validateExecutionClaimTerminalBudget([]json.RawMessage{raw}); err == nil {
		t.Fatal("invalid revision was accepted as a bounded security condition")
	}
}
