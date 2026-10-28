package riidoaiserver

import "context"

func (s *DevelopmentAIAgentClientStore) cliInstallationDeviceOwner(ctx context.Context, id string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deviceCredentials[id].ownerPrincipalID, nil
}

func (s *PersistentAIAgentClientStore) cliInstallationDeviceOwner(ctx context.Context, id string) (string, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return "", err
	}
	return s.DevelopmentAIAgentClientStore.cliInstallationDeviceOwner(ctx, id)
}
