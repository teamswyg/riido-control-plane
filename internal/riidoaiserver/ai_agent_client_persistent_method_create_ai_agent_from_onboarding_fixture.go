package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) CreateAIAgentFromOnboardingFixture(ctx context.Context, principal AuthorizationResult, fixtureID string, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientRecordResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgentFromOnboardingFixture(ctx, principal, fixtureID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
