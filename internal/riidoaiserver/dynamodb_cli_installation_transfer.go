package riidoaiserver

import (
	"context"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) confirmCLIInstallationTransfer(ctx context.Context, id string, device authenticatedClaimDevice, signature, resource string) (cliInstallationLink, error) {
	for range 3 {
		if err := ctx.Err(); err != nil {
			return cliInstallationLink{}, err
		}
		intent, oldIntent, err := s.loadCLIInstallationIntent(ctx, id)
		if err != nil {
			return cliInstallationLink{}, err
		}
		if !intent.Request.isTransfer() || intent.Request.Resource != resource {
			return cliInstallationLink{}, ErrAuthorizationForbidden
		}
		if err := verifyCLIInstallationProof(intent, device, signature); err != nil {
			return cliInstallationLink{}, err
		}
		if intent.Retired || (!intent.Consumed && !s.now().Before(intent.ExpiresAt)) {
			return cliInstallationLink{}, errCLIInstallationExpired
		}
		link, oldLink, err := s.loadCLIInstallationLink(ctx, intent.Request.cliInstallationIdentity)
		if errors.Is(err, ErrAuthorizationForbidden) || (err == nil && len(oldLink) == 0) {
			return cliInstallationLink{}, errCLIInstallationConflict
		}
		if err != nil {
			return cliInstallationLink{}, err
		}
		if err := checkCLIInstallationTransferLink(intent, link, device); err != nil {
			return cliInstallationLink{}, err
		}
		source, target, err := s.cliInstallationTransferCoverage(ctx, intent, device)
		if err != nil {
			return cliInstallationLink{}, err
		}
		if !intent.Consumed {
			now := s.now()
			if !now.Before(intent.ExpiresAt) {
				return cliInstallationLink{}, errCLIInstallationExpired
			}
			link = linkForCLIInstallation(intent, device, now)
		}
		actions, err := s.cliInstallationTransferActions(intent, link, oldIntent, oldLink, source, target)
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
