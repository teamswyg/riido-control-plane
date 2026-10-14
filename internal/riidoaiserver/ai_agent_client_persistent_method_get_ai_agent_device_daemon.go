package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) GetAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string) (DeviceDaemonDetailResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonDetailResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentDeviceDaemon(ctx, principal, deviceID)
}
