package riidoaiserver

import (
	"math"
	"strconv"
)

const executionCoverageCAS = "policy_schema = :schema AND generation = :generation AND credential_generation = :credential AND owner_principal_id = :owner AND coverage_state = :state AND revision = :revision AND unresolved = :unresolved"

func executionCoverageCondition(table string, c executionCoverage) *dynamoDBTransactWriteConditionCheckAction {
	return &dynamoDBTransactWriteConditionCheckAction{TableName: table, Key: devicePolicyKey(c.DeviceID), ConditionExpression: executionCoverageCAS, ExpressionAttributeValues: map[string]map[string]string{
		":schema": {"S": executionCoverageSchema}, ":generation": {"S": c.Generation},
		":credential": {"S": c.CredentialGeneration}, ":owner": {"S": c.OwnerPrincipalID},
		":state": {"S": c.State}, ":revision": {"N": strconv.FormatUint(c.Revision, 10)},
		":unresolved": {"N": strconv.FormatUint(c.Unresolved, 10)},
	}}
}

// The same bounded CAS is used by enrollment rotation, claims, and termination.
// An empty nextCredential preserves the credential generation; delta is -1/0/1.
func executionCoverageChangeAction(table string, before executionCoverage, delta int, nextCredential string) (dynamoDBTransactWriteItem, error) {
	if _, err := executionCoverageItem(before); err != nil || before.Revision == math.MaxUint64 || delta < -1 || delta > 1 || (delta < 0 && before.Unresolved == 0) || (delta > 0 && before.Unresolved == math.MaxUint64) {
		return dynamoDBTransactWriteItem{}, errCLIInstallationUnavailable
	}
	after := before
	after.Revision++
	switch delta {
	case -1:
		after.Unresolved--
	case 1:
		after.Unresolved++
	}
	if nextCredential != "" {
		after.CredentialGeneration = nextCredential
	}
	item, err := executionCoverageItem(after)
	if err != nil {
		return dynamoDBTransactWriteItem{}, err
	}
	condition := executionCoverageCondition(table, before)
	return dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: table, Item: item, ConditionExpression: condition.ConditionExpression, ExpressionAttributeValues: condition.ExpressionAttributeValues}}, nil
}
