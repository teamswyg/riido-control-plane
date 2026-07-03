package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentDeviceDaemons(ctx context.Context, principal AuthorizationResult, deviceID string) (DeviceDaemonListResponse, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentDeviceDaemons(ctx, principal, deviceID)
}
