package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentDeviceDaemons(ctx context.Context, principal AuthorizationResult, deviceID string) (DeviceDaemonListResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentDeviceDaemons(ctx, principal, deviceID)
}
