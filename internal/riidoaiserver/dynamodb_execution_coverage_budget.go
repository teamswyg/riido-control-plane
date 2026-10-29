package riidoaiserver

import (
	"reflect"
	"strings"
)

func legacyExecutionAdmissionFence(table, device string) dynamoDBTransactWriteItem {
	return dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: table, Key: devicePolicyKey(device), ConditionExpression: "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR (policy_schema = :schema AND coverage_state = :unknown)", ExpressionAttributeValues: map[string]map[string]string{":schema": {"S": devicePolicyRevisionSchema}, ":unknown": {"S": "unknown"}}}}
}

func isBoundedCoverageClaimAction(action dynamoDBTransactWriteItem) bool {
	if check := action.ConditionCheck; check != nil {
		device := strings.TrimPrefix(dynamoDBStringValue(check.Key, "pk"), "DEVICE_POLICY#")
		return cliInstallationID(device) && (reflect.DeepEqual(action, legacyExecutionAdmissionFence(check.TableName, device)) || boundedCoverageCondition(action))
	}
	put := action.Put
	if put == nil || put.ConditionExpression != executionCoverageCAS {
		return false
	}
	device := strings.TrimPrefix(dynamoDBStringValue(put.Item, "pk"), "DEVICE_POLICY#")
	after, err := executionCoverageFromItem(put.Item, device)
	if err != nil {
		return false
	}
	before := after
	if before.Revision, err = coverageCounter(put.ExpressionAttributeValues[":revision"]["N"]); err != nil {
		return false
	}
	if before.Unresolved, err = coverageCounter(put.ExpressionAttributeValues[":unresolved"]["N"]); err != nil {
		return false
	}
	for _, delta := range []int{0, 1} {
		expected, err := executionCoverageChangeAction(put.TableName, before, delta, "")
		if err == nil && reflect.DeepEqual(action, expected) {
			return true
		}
	}
	return false
}

func devicePolicyActionKey(action dynamoDBTransactWriteItem) string {
	switch {
	case action.Update != nil:
		return dynamoDBStringValue(action.Update.Key, "pk")
	case action.Put != nil:
		return dynamoDBStringValue(action.Put.Item, "pk")
	case action.ConditionCheck != nil:
		return dynamoDBStringValue(action.ConditionCheck.Key, "pk")
	default:
		return ""
	}
}
