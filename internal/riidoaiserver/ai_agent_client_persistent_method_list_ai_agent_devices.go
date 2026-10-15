package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceRuntimeListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentDevices(ctx, principal)
}
