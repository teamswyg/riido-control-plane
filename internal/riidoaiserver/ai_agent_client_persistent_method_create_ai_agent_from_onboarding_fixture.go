package riidoaiserver

import (
	"context"
	"errors"
)

func (s *PersistentAIAgentClientStore) CreateAIAgentFromOnboardingFixture(ctx context.Context, principal AuthorizationResult, fixtureID string, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	for attempt := 0; attempt < 3; attempt++ {
		s.requireOnboardingCreateReload()
		if err := s.reloadSnapshot(ctx); err != nil {
			return AgentClientRecordResponse{}, onboardingAgentCreateDependencyError{cause: err}
		}
		trial, before, err := s.enrollmentTrial()
		if err != nil {
			return AgentClientRecordResponse{}, onboardingAgentCreateDependencyError{cause: err}
		}
		response, err := trial.CreateAIAgentFromOnboardingFixture(ctx, principal, fixtureID, req)
		if err != nil {
			return response, err
		}
		// The forced fresh load already proves this replay durable; no new save is needed.
		for _, agent := range before.Agents {
			if agent.AgentID == response.Agent.AgentID {
				return response, nil
			}
		}
		snapshot, err := trial.snapshot(s.snapshotNow().UTC())
		if err != nil {
			return AgentClientRecordResponse{}, onboardingAgentCreateDependencyError{cause: err}
		}
		snapshot.ExpectedVersion, snapshot.Version = s.version, s.version+1
		writeCtx := context.WithValue(ctx, taskAgentPreserveMembershipContextKey{}, true)
		writeCtx = context.WithValue(writeCtx, onboardingAgentCreateMutationContextKey{}, true)
		if err := s.snapshotStore.SaveAIAgentClientSnapshot(writeCtx, snapshot); err != nil {
			if committed, ok := s.resolveOnboardingAgentCreate(ctx, principal, snapshot, response); ok {
				return committed, nil
			}
			if errors.Is(err, errTaskAgentSnapshotChanged) || errors.Is(err, errAIAgentSnapshotConflict) {
				continue
			}
			return AgentClientRecordResponse{}, onboardingAgentCreateDependencyError{cause: err}
		}
		s.publishOnboardingAgentCreate(snapshot, response.Agent.AgentID)
		s.version = snapshot.Version
		s.markSnapshotSaved(snapshot)
		return response, nil
	}
	return AgentClientRecordResponse{}, onboardingAgentCreateDependencyError{cause: errTaskAgentSnapshotChanged}
}
