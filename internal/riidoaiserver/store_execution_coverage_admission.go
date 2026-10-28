package riidoaiserver

import (
	"context"
	"strings"
)

type executionCoverageAdmissionReader interface {
	loadExecutionCoverage(context.Context, string) (executionCoverage, bool, error)
	authorizeCoveredAssignment(context.Context, executionCoverage, Assignment) error
}

func (s *Store) prepareExecutionAdmission(ctx context.Context, agentID string, req PollRequest) (context.Context, error) {
	deviceID := ""
	device, authenticated := claimDeviceFromContext(ctx)
	if authenticated {
		deviceID = device.deviceID
	} else if s.agentRegistry != nil {
		if binding, found := lookupDaemonBinding(s.agentRegistry, agentID); found {
			deviceID = binding.DeviceID
		}
	}
	if deviceID == "" {
		if strings.HasPrefix(req.DeviceID, executionCoveragePrefix) {
			return ctx, ErrAuthorizationForbidden
		}
		if _, supported := s.operationStore.(executionCoverageAdmissionReader); supported {
			return ctx, ErrAuthorizationForbidden
		}
		return ctx, nil
	}
	reader, supported := s.operationStore.(executionCoverageAdmissionReader)
	if !supported {
		if strings.HasPrefix(deviceID, executionCoveragePrefix) {
			return ctx, errCLIInstallationUnavailable
		}
		return ctx, nil
	}
	coverage, covered, err := reader.loadExecutionCoverage(ctx, deviceID)
	if err != nil {
		return ctx, err
	}
	admission := &executionAdmission{deviceID: deviceID}
	if covered {
		auth, ok := ctx.Value(executionCoverageAuthKey{}).(executionCoverageAuth)
		if !authenticated || !ok || coverage.State != "complete" || coverage.OwnerPrincipalID != device.ownerPrincipalID || coverage.Generation != auth.generation || coverage.CredentialGeneration != auth.credential {
			return ctx, ErrAuthorizationForbidden
		}
		admission.coverage = &coverage
	}
	return context.WithValue(ctx, executionAdmissionKey{}, admission), nil
}

func (s *Store) authorizeCoveredAssignment(ctx context.Context, assignment Assignment) error {
	admission := executionAdmissionFromContext(ctx)
	if admission == nil || admission.coverage == nil {
		return nil
	}
	reader, ok := s.operationStore.(executionCoverageAdmissionReader)
	if !ok {
		return errCLIInstallationUnavailable
	}
	return reader.authorizeCoveredAssignment(ctx, *admission.coverage, assignment)
}
