package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) GetAIAgentEditability(ctx context.Context, principal AuthorizationResult, agentID string) (AgentEditabilityResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentEditabilityResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.GetAIAgentEditability(ctx, principal, agentID)
}
