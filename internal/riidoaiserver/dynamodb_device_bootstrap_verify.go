package riidoaiserver

import (
	"context"
	"reflect"
)

func (s *DynamoDBAIAgentClientSnapshot) verifyBootstrapState(ctx context.Context, r deviceBootstrapRecord) error {
	credentials, err := s.desktopPresenceCredentials(ctx)
	if err != nil {
		return err
	}
	item, err := s.bootstrapGet(ctx, devicePolicyKey(r.Before.DeviceID), credentials)
	if err != nil {
		return err
	}
	switch r.State {
	case "issued":
		if len(item) != 0 && dynamoDBStringValue(item, "policy_schema") != devicePolicyRevisionSchema {
			return errCLIInstallationConflict
		}
		_, err := devicePolicyRevision(item, r.Before.DeviceID)
		return err
	case "prepared":
		if !reflect.DeepEqual(item, bootstrapFenceItem(r)) {
			return errCLIInstallationConflict
		}
	case "complete":
		coverage, err := executionCoverageFromItem(item, r.After.DeviceID)
		if err != nil {
			return err
		}
		if coverage.Generation != r.After.ExecutionCoverageGeneration || coverage.CredentialGeneration != r.After.CredentialGeneration || coverage.OwnerPrincipalID != r.After.OwnerPrincipalID || coverage.State == "revoked" || coverage.Revision < r.PolicyRevision {
			return errCLIInstallationConflict
		}
	default:
		return errCLIInstallationInvalid
	}
	return nil
}
