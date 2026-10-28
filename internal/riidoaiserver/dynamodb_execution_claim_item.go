package riidoaiserver

import (
	"encoding/json"
	"errors"
)

const (
	executionClaimSchema        = "execution_claim/v1"
	executionClaimCoveredSchema = "execution_claim/v2"
	executionClaimItemBytes     = 4 << 10
)

func executionClaimItem(claim executionClaim) (map[string]map[string]string, error) {
	if err := claim.validate(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(claim)
	if err != nil {
		return nil, err
	}
	item := map[string]map[string]string{
		"pk": {"S": "DEVICE_EXECUTION#" + claim.DeviceID}, "sk": {"S": claim.key()},
		"schema": {"S": executionClaimSchema}, "payload": {"S": string(payload)},
	}
	if claim.CoverageGeneration != "" {
		item["schema"]["S"] = executionClaimCoveredSchema
	}
	encoded, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	if len(encoded) > executionClaimItemBytes-512 {
		return nil, errors.New("execution claim byte budget exceeded")
	}
	return item, nil
}

func executionClaimFromItem(item map[string]map[string]string) (executionClaim, error) {
	var claim executionClaim
	if err := json.Unmarshal([]byte(dynamoDBStringValue(item, "payload")), &claim); err != nil {
		return claim, err
	}
	if err := claim.validate(); err != nil {
		return claim, err
	}
	schema := executionClaimSchema
	if claim.CoverageGeneration != "" {
		schema = executionClaimCoveredSchema
	}
	if dynamoDBStringValue(item, "schema") != schema || dynamoDBStringValue(item, "pk") != "DEVICE_EXECUTION#"+claim.DeviceID || dynamoDBStringValue(item, "sk") != claim.key() {
		return claim, errors.New("execution claim storage scope mismatch")
	}
	return claim, nil
}

func (s *DynamoDBAssignmentOperationStore) appendExecutionClaimAction(items []dynamoDBTransactWriteItem, claim *executionClaim, admissions ...*executionAdmission) ([]dynamoDBTransactWriteItem, error) {
	if claim == nil {
		return s.appendExecutionAdmissionActions(items, nil, admissions...)
	}
	item, err := executionClaimItem(*claim)
	if err != nil {
		return nil, err
	}
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{
		TableName: s.tableName, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)", Item: item,
	}})
	return s.appendExecutionAdmissionActions(items, claim, admissions...)
}
