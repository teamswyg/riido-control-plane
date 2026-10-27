package riidoaiserver

import (
	"math"
	"strconv"
)

const devicePolicyRevisionSchema = "device_policy_revision/v1"

func devicePolicyKey(device string) map[string]map[string]string {
	return map[string]map[string]string{"pk": {"S": "DEVICE_POLICY#" + device}, "sk": {"S": "REVISION"}}
}

func (s *DynamoDBAssignmentOperationStore) devicePolicyAdvanceAction(device string) dynamoDBTransactWriteItem {
	return dynamoDBTransactWriteItem{Update: &dynamoDBTransactWriteUpdateAction{
		TableName: s.tableName, Key: devicePolicyKey(device),
		UpdateExpression:          "SET policy_schema = if_not_exists(policy_schema, :schema), coverage_state = if_not_exists(coverage_state, :unknown), revision = if_not_exists(revision, :zero) + :one",
		ConditionExpression:       "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR (policy_schema = :schema AND coverage_state = :unknown AND revision > :zero AND revision < :max)",
		ExpressionAttributeValues: map[string]map[string]string{":schema": {"S": devicePolicyRevisionSchema}, ":unknown": {"S": "unknown"}, ":zero": {"N": "0"}, ":one": {"N": "1"}, ":max": {"N": strconv.FormatUint(math.MaxUint64, 10)}},
	}}
}

func devicePolicyRevision(item map[string]map[string]string, device string) (uint64, error) {
	if len(item) == 0 {
		return 0, nil
	}
	if dynamoDBStringValue(item, "pk") != "DEVICE_POLICY#"+device || dynamoDBStringValue(item, "sk") != "REVISION" || dynamoDBStringValue(item, "policy_schema") != devicePolicyRevisionSchema || dynamoDBStringValue(item, "coverage_state") != "unknown" {
		return 0, errCLIInstallationUnavailable
	}
	value := item["revision"]["N"]
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil || n == 0 || strconv.FormatUint(n, 10) != value {
		return 0, errCLIInstallationUnavailable
	}
	return n, nil
}

func (s *DynamoDBAssignmentOperationStore) devicePolicyFence(device string, revision uint64) dynamoDBTransactWriteItem {
	return dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: devicePolicyKey(device), ConditionExpression: "policy_schema = :schema AND coverage_state = :unknown AND revision = :revision", ExpressionAttributeValues: map[string]map[string]string{":schema": {"S": devicePolicyRevisionSchema}, ":unknown": {"S": "unknown"}, ":revision": {"N": strconv.FormatUint(revision, 10)}}}}
}
