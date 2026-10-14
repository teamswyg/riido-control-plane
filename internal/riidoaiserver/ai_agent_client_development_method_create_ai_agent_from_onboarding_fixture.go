package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) CreateAIAgentFromOnboardingFixture(ctx context.Context, principal AuthorizationResult, fixtureID string, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentClientRecordResponse{}, err
	}
	fixtureID = strings.TrimSpace(fixtureID)
	if fixtureID == "" {
		return AgentClientRecordResponse{}, errors.New("fixture_id is required")
	}
	s.mu.Lock()
	s.ensureOnboardingFixtureColorsLocked()
	fixture, ok := s.findOnboardingFixtureLocked(fixtureID)
	s.mu.Unlock()
	if !ok {
		return AgentClientRecordResponse{}, ErrAIAgentNotFound
	}
	return s.createAIAgent(ctx, principal, req, fixture.TmpColor)
}
