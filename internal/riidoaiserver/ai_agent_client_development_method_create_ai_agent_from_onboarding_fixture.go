package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) CreateAIAgentFromOnboardingFixture(ctx context.Context, principal AuthorizationResult, fixtureID string, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	s.onboardingCreateMu.Lock()
	defer s.onboardingCreateMu.Unlock()
	if err := ctx.Err(); err != nil {
		return AgentClientRecordResponse{}, err
	}
	fixtureID = strings.TrimSpace(fixtureID)
	if fixtureID == "" {
		return AgentClientRecordResponse{}, errors.New("fixture_id is required")
	}
	s.mu.Lock()
	key, fingerprint, err := onboardingAgentCreateIdentity(ctx, principal, s.workspaceScope(principal), fixtureID, req)
	if err != nil {
		s.mu.Unlock()
		return AgentClientRecordResponse{}, err
	}
	now := time.Now().UTC()
	if key != "" {
		response, found, replayErr := s.replayOnboardingAgentCreateLocked(principal, key, fingerprint, now)
		if found {
			s.mu.Unlock()
			return response, replayErr
		}
	}
	s.ensureOnboardingFixtureColorsLocked()
	fixture, found := s.findOnboardingFixtureLocked(fixtureID)
	s.mu.Unlock()
	if !found {
		return AgentClientRecordResponse{}, ErrAIAgentNotFound
	}
	response, err := s.createAIAgentWithProfileColor(ctx, principal, req, fixture.TmpColor, onboardingAgentProfileColor, true)
	if err != nil || key == "" {
		return response, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	receipts := retainOnboardingAgentCreates(s.onboardingAgentCreates, now)
	receipts[key] = onboardingAgentCreateReceipt{Fingerprint: fingerprint, AgentID: response.Agent.AgentID, AgentCreatedAt: response.Agent.CreatedAt, CreatedAt: now}
	raw, err := json.Marshal(receipts)
	if err != nil || len(raw) > onboardingAgentCreateReceiptBytes {
		delete(s.agents, response.Agent.AgentID)
		delete(s.onboardingCreated, response.Agent.AgentID)
		return AgentClientRecordResponse{}, onboardingAgentCreateDependencyError{cause: errors.New("onboarding receipt capacity exhausted")}
	}
	s.onboardingAgentCreates = receipts
	return response, nil
}
