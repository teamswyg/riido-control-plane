package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) authorizeCoveredAssignment(ctx context.Context, coverage executionCoverage, a Assignment) error {
	key := a.ID + "#" + leaseEpoch(a.LeaseToken)
	item, err := s.executionTerminationGet(ctx, "DEVICE_EXECUTION#"+coverage.DeviceID, key)
	if err != nil {
		return err
	}
	claim, err := executionClaimFromItem(item)
	if err != nil || claim.CoverageGeneration != coverage.Generation || claim.OwnerPrincipalID != coverage.OwnerPrincipalID || claim.AgentID != a.AgentID {
		return ErrAuthorizationForbidden
	}
	receipt, err := s.executionTerminationGet(ctx, "DEVICE_EXECUTION_TERMINATION#"+coverage.DeviceID, key)
	if err != nil {
		return err
	}
	if len(receipt) > 0 || isTerminal(a.State) {
		return ErrAuthorizationForbidden
	}
	// Final authority fence also protects a read-only PollActive/duplicate Ready.
	_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{[]dynamoDBTransactWriteItem{
		{ConditionCheck: executionCoverageCondition(s.tableName, coverage)},
		{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: map[string]map[string]string{"pk": {"S": "DEVICE_EXECUTION_TERMINATION#" + coverage.DeviceID}, "sk": {"S": key}}, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}},
		{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: map[string]map[string]string{"pk": {"S": dynamoDBAssignmentProjectionPK(a.ID)}, "sk": {"S": dynamoDBAssignmentProjectionSK}}, ConditionExpression: "schema_version = :schema AND assignment_state = :state AND lease_epoch = :epoch", ExpressionAttributeValues: map[string]map[string]string{":schema": {"S": AssignmentProjectionSchemaVersion}, ":state": {"S": string(a.State)}, ":epoch": {"S": leaseEpoch(a.LeaseToken)}}}},
	}})
	return err
}
