package riidoaiserver

import "context"

type (
	cliInstallationCoverageAuthority       struct{ owner, generation, credential string }
	cliInstallationCoverageAuthorityReader interface {
		cliInstallationDeviceCoverage(context.Context, string) (cliInstallationCoverageAuthority, error)
	}
)

func (s *DevelopmentAIAgentClientStore) cliInstallationDeviceCoverage(ctx context.Context, id string) (cliInstallationCoverageAuthority, error) {
	if err := ctx.Err(); err != nil {
		return cliInstallationCoverageAuthority{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.deviceCredentials[id]
	return cliInstallationCoverageAuthority{c.ownerPrincipalID, c.coverageGeneration, c.credentialGeneration}, nil
}

func (s *PersistentAIAgentClientStore) cliInstallationDeviceCoverage(ctx context.Context, id string) (cliInstallationCoverageAuthority, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return cliInstallationCoverageAuthority{}, err
	}
	return s.DevelopmentAIAgentClientStore.cliInstallationDeviceCoverage(ctx, id)
}

func currentCLIInstallationCoverage(ctx context.Context, owners cliInstallationDeviceOwnerReader, coverage executionCoverage) (bool, error) {
	reader, ok := owners.(cliInstallationCoverageAuthorityReader)
	if !ok {
		return false, nil
	}
	authority, err := reader.cliInstallationDeviceCoverage(ctx, coverage.DeviceID)
	return err == nil && authority.owner == coverage.OwnerPrincipalID && authority.generation == coverage.Generation && authority.credential == coverage.CredentialGeneration, err
}
