package riidoaiserver

func (s *DynamoDBAssignmentOperationStore) appendCLIInstallationCoverage(actions []dynamoDBTransactWriteItem, coverage *executionCoverage, advance bool) ([]dynamoDBTransactWriteItem, error) {
	if coverage == nil {
		return actions, nil
	}
	if !advance {
		return append(actions, dynamoDBTransactWriteItem{ConditionCheck: executionCoverageCondition(s.tableName, *coverage)}), nil
	}
	action, err := executionCoverageChangeAction(s.tableName, *coverage, 0, "")
	return append(actions, action), err
}
