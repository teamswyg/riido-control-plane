package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) UpdateAIAgentConfiguration(ctx context.Context, principal AuthorizationResult, agentID string, req UpdateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientRecordResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.UpdateAIAgentConfiguration(ctx, principal, agentID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
