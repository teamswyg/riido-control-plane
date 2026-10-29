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
		now := s.now()
		if !intent.Consumed && !now.Before(intent.ExpiresAt) {
			return cliInstallationLink{}, errCLIInstallationExpired
		}
		if len(oldLink) == 0 {
			link = linkForCLIInstallation(intent, device, now)
		}
		coverage, covered, err := s.loadExecutionCoverage(ctx, device.deviceID)
		if err != nil {
			return cliInstallationLink{}, err
		}
		var authority *executionCoverage
		if covered {
			if coverage.OwnerPrincipalID != device.ownerPrincipalID || coverage.Generation != device.coverage.generation || coverage.CredentialGeneration != device.coverage.credential {
				return cliInstallationLink{}, ErrAuthorizationForbidden
			}
			authority = &coverage
		}
		actions, err := s.cliInstallationConfirmActions(intent, link, oldIntent, oldLink, authority)
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
