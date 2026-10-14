package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) GetAIAgentDaemon(ctx context.Context, principal AuthorizationResult, agentID string) (DeviceDaemonDetailResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonDetailResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentDaemon(ctx, principal, agentID)
}
