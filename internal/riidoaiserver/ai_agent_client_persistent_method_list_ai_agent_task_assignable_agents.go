package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error) {
	if err := s.lockOperationContext(ctx); err != nil {
		return AgentClientListResponse{}, err
	}
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
}
