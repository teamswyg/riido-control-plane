package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) CreateAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientRecordResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgent(ctx, principal, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
