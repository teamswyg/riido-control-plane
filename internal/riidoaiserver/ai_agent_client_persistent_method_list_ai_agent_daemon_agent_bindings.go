package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentDaemonAgentBindings(ctx context.Context, principal AuthorizationResult, deviceID string) (AgentRuntimeBindingListResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentRuntimeBindingListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentDaemonAgentBindings(ctx, principal, deviceID)
}
