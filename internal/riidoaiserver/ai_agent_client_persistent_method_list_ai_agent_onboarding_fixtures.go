package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentOnboardingFixtures(ctx context.Context, principal AuthorizationResult) (AgentOnboardingFixtureListResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentOnboardingFixtureListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentOnboardingFixtures(ctx, principal)
}
