package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
}
