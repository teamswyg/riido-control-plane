package riidoaiserver

import (
	"context"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) confirmCLIInstallationLink(ctx context.Context, id string, device authenticatedClaimDevice, signature, resource string) (cliInstallationLink, error) {
	for range 3 {
		if err := ctx.Err(); err != nil {
			return cliInstallationLink{}, err
		}
		intent, oldIntent, err := s.loadCLIInstallationIntent(ctx, id)
		if err != nil {
			return cliInstallationLink{}, err
		}
		if intent.Request.Resource != resource {
			return cliInstallationLink{}, ErrAuthorizationForbidden
		}
		if err = verifyCLIInstallationProof(intent, device, signature); err != nil {
			return cliInstallationLink{}, err
		}
		link, oldLink, err := s.loadCLIInstallationLink(ctx, intent.Request.cliInstallationIdentity)
		if errors.Is(err, ErrAuthorizationForbidden) {
			return cliInstallationLink{}, errCLIInstallationConflict
		}
		if err != nil {
			return cliInstallationLink{}, err
		}
		if len(oldLink) > 0 && !matchingCLIInstallationLink(link, intent, device) {
			return cliInstallationLink{}, errCLIInstallationConflict
		}
		if intent.Consumed && len(oldLink) == 0 {
			return cliInstallationLink{}, errCLIInstallationUnavailable
		}
		if !intent.Consumed && !s.now().Before(intent.ExpiresAt) {
			return cliInstallationLink{}, errCLIInstallationExpired
		}
		if len(oldLink) == 0 {
			link = linkForCLIInstallation(intent, device, s.now())
		}
		actions, err := s.cliInstallationConfirmActions(intent, link, oldIntent, oldLink)
		if err != nil {
			return cliInstallationLink{}, err
		}
		_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
		if err == nil {
			return link, nil
		}
	}
	return cliInstallationLink{}, errCLIInstallationUnavailable
}

func (s *DynamoDBAssignmentOperationStore) cliInstallationConfirmActions(intent cliInstallationIntent, link cliInstallationLink, oldIntent, oldLink map[string]map[string]string) ([]dynamoDBTransactWriteItem, error) {
	if intent.Consumed {
		return []dynamoDBTransactWriteItem{{ConditionCheck: s.installationItemCondition(oldIntent, cliInstallationIntentKey(intent.ID))}, {ConditionCheck: s.installationItemCondition(oldLink, cliInstallationLinkKey(link.cliInstallationIdentity))}}, nil
	}
	intent.Consumed = true
	newIntent, err := cliInstallationItem(cliInstallationIntentKey(intent.ID), cliInstallationIntentItemSchema, intent)
	if err != nil {
		return nil, err
	}
	actions := []dynamoDBTransactWriteItem{s.installationItemPut(newIntent, oldIntent)}
	if len(oldLink) > 0 {
		return append(actions, dynamoDBTransactWriteItem{ConditionCheck: s.installationItemCondition(oldLink, cliInstallationLinkKey(link.cliInstallationIdentity))}), nil
	}
	newLink, err := cliInstallationItem(cliInstallationLinkKey(link.cliInstallationIdentity), cliInstallationLinkItemSchema, link)
	if err != nil {
		return nil, err
	}
	return append(actions, s.installationItemPut(newLink, nil), s.devicePolicyAdvanceAction(link.DeviceID)), nil
}
