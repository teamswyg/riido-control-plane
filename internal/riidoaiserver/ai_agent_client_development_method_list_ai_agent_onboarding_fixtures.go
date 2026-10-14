package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) ListAIAgentOnboardingFixtures(ctx context.Context, principal AuthorizationResult) (AgentOnboardingFixtureListResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentOnboardingFixtureListResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureOnboardingFixtureColorsLocked()
	return AgentOnboardingFixtureListResponse{
		SchemaVersion: SchemaVersion,
		Fixtures:      copyAgentOnboardingFixtures(s.fixtures),
	}, nil
}
