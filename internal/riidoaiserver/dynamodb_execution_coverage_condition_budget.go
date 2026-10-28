package riidoaiserver

import (
	"reflect"
	"strings"
)

func boundedCoverageCondition(action dynamoDBTransactWriteItem) bool {
	check := action.ConditionCheck
	if check == nil || check.ConditionExpression != executionCoverageCAS {
		return false
	}
	item := map[string]map[string]string{"pk": check.Key["pk"], "sk": check.Key["sk"]}
	for field, variable := range map[string]string{"policy_schema": ":schema", "generation": ":generation", "credential_generation": ":credential", "owner_principal_id": ":owner", "coverage_state": ":state", "revision": ":revision", "unresolved": ":unresolved"} {
		item[field] = check.ExpressionAttributeValues[variable]
	}
	device := strings.TrimPrefix(dynamoDBStringValue(check.Key, "pk"), "DEVICE_POLICY#")
	coverage, err := executionCoverageFromItem(item, device)
	return err == nil && reflect.DeepEqual(action, dynamoDBTransactWriteItem{ConditionCheck: executionCoverageCondition(check.TableName, coverage)})
}
