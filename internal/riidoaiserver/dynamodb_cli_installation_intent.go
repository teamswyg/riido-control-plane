package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) createCLIInstallationIntent(ctx context.Context, intent cliInstallationIntent) error {
	if err := intent.valid(); err != nil {
		return err
	}
	item, err := cliInstallationItem(cliInstallationIntentKey(intent.ID), cliInstallationIntentItemSchema, intent)
	if err != nil {
		return err
	}
	_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{[]dynamoDBTransactWriteItem{s.installationItemPut(item, nil)}})
	return err
}

func (s *DynamoDBAssignmentOperationStore) loadCLIInstallationIntent(ctx context.Context, id string) (cliInstallationIntent, map[string]map[string]string, error) {
	var intent cliInstallationIntent
	key := cliInstallationIntentKey(id)
	item, err := s.executionTerminationGet(ctx, key["pk"]["S"], key["sk"]["S"])
	if err != nil {
		return intent, nil, err
	}
	if len(item) == 0 {
		return intent, nil, ErrAuthorizationForbidden
	}
	if err = decodeCLIInstallationItem(item, key, cliInstallationIntentItemSchema, &intent); err != nil {
		return intent, nil, err
	}
	if intent.ID != id || intent.valid() != nil {
		return intent, nil, errCLIInstallationUnavailable
	}
	return intent, item, nil
}

func (s *DynamoDBAssignmentOperationStore) loadCLIInstallationLink(ctx context.Context, id cliInstallationIdentity) (cliInstallationLink, map[string]map[string]string, error) {
	var link cliInstallationLink
	key := cliInstallationLinkKey(id)
	item, err := s.executionTerminationGet(ctx, key["pk"]["S"], key["sk"]["S"])
	if err != nil || len(item) == 0 {
		return link, item, err
	}
	if err = decodeCLIInstallationItem(item, key, cliInstallationLinkItemSchema, &link); err != nil {
		return link, nil, err
	}
	if link.cliInstallationIdentity != id {
		return link, nil, ErrAuthorizationForbidden
	}
	if err = link.valid(); err != nil {
		return link, nil, errCLIInstallationUnavailable
	}
	return link, item, nil
}
