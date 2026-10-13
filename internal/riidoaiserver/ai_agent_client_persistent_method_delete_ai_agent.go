package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) DeleteAIAgent(ctx context.Context, principal AuthorizationResult, agentID string) (DeleteAgentResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeleteAgentResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.DeleteAIAgent(ctx, principal, agentID)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
