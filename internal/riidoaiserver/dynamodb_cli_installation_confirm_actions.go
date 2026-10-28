package riidoaiserver

func (s *DynamoDBAssignmentOperationStore) cliInstallationConfirmActions(intent cliInstallationIntent, link cliInstallationLink, oldIntent, oldLink map[string]map[string]string, authorities ...*executionCoverage) ([]dynamoDBTransactWriteItem, error) {
	var coverage *executionCoverage
	if len(authorities) > 0 {
		coverage = authorities[0]
	}
	if intent.Consumed {
		actions := []dynamoDBTransactWriteItem{{ConditionCheck: s.installationItemCondition(oldIntent, cliInstallationIntentKey(intent.ID))}, {ConditionCheck: s.installationItemCondition(oldLink, cliInstallationLinkKey(link.cliInstallationIdentity))}}
		return s.appendCLIInstallationCoverage(actions, coverage, false)
	}
	intent.Consumed = true
	newIntent, err := cliInstallationItem(cliInstallationIntentKey(intent.ID), cliInstallationIntentItemSchema, intent)
	if err != nil {
		return nil, err
	}
	actions := []dynamoDBTransactWriteItem{s.installationItemPut(newIntent, oldIntent)}
	if len(oldLink) > 0 {
		actions = append(actions, dynamoDBTransactWriteItem{ConditionCheck: s.installationItemCondition(oldLink, cliInstallationLinkKey(link.cliInstallationIdentity))})
		return s.appendCLIInstallationCoverage(actions, coverage, false)
	}
	newLink, err := cliInstallationItem(cliInstallationLinkKey(link.cliInstallationIdentity), cliInstallationLinkItemSchema, link)
	if err != nil {
		return nil, err
	}
	actions = append(actions, s.installationItemPut(newLink, nil))
	if coverage != nil {
		return s.appendCLIInstallationCoverage(actions, coverage, true)
	}
	return append(actions, s.devicePolicyAdvanceAction(link.DeviceID)), nil
}
