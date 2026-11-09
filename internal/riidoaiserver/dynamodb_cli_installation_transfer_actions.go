package riidoaiserver

func (s *DynamoDBAssignmentOperationStore) cliInstallationTransferActions(intent cliInstallationIntent, link cliInstallationLink, oldIntent, oldLink map[string]map[string]string, source, target executionCoverage) ([]dynamoDBTransactWriteItem, error) {
	if intent.Consumed {
		return []dynamoDBTransactWriteItem{
			{ConditionCheck: s.installationItemCondition(oldIntent, cliInstallationIntentKey(intent.ID))},
			{ConditionCheck: s.installationItemCondition(oldLink, cliInstallationLinkKey(link.cliInstallationIdentity))},
			{ConditionCheck: executionCoverageCondition(s.tableName, source)},
			{ConditionCheck: executionCoverageCondition(s.tableName, target)},
		}, nil
	}
	// All admission paths for covered devices require complete coverage. The
	// shared CAS also defeats an in-flight claim or credential rotation.
	retire, err := executionCoverageChangeAction(s.tableName, source, 0, "")
	if err != nil {
		return nil, err
	}
	retire.Put.Item["coverage_state"] = map[string]string{"S": "revoked"}
	intent.Consumed = true
	newIntent, err := cliInstallationItem(cliInstallationIntentKey(intent.ID), cliInstallationIntentItemSchema, intent)
	if err != nil {
		return nil, err
	}
	newLink, err := cliInstallationItem(cliInstallationLinkKey(link.cliInstallationIdentity), cliInstallationLinkItemSchema, link)
	if err != nil {
		return nil, err
	}
	return []dynamoDBTransactWriteItem{
		s.installationItemPut(newIntent, oldIntent),
		s.installationItemPut(newLink, oldLink),
		retire,
		{ConditionCheck: executionCoverageCondition(s.tableName, target)},
	}, nil
}
