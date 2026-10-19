package riidoaiserver

import (
	"context"
	"time"
)

// A transport timeout is not proof that a transaction failed. Only an exact
// durable receipt (or exact unkeyed record) permits publishing an uncertain save.
func (s *PersistentAIAgentClientStore) resolveOnboardingAgentCreate(ctx context.Context, principal AuthorizationResult, proposed AIAgentClientSnapshot, response AgentClientRecordResponse) (AgentClientRecordResponse, bool) {
	s.requireOnboardingCreateReload()
	readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	committed, found, err := s.snapshotStore.LoadAIAgentClientSnapshot(readCtx)
	if err != nil || !found {
		return AgentClientRecordResponse{}, false
	}
	key := ""
	for candidate, receipt := range proposed.OnboardingAgentCreates {
		if receipt.AgentID == response.Agent.AgentID && receipt.AgentCreatedAt.Equal(response.Agent.CreatedAt) {
			key = candidate
			break
		}
	}
	if key != "" {
		receipt, ok := committed.OnboardingAgentCreates[key]
		expected := proposed.OnboardingAgentCreates[key]
		if !ok || receipt.Fingerprint != expected.Fingerprint || receipt.AgentID != expected.AgentID || !receipt.CreatedAt.Equal(expected.CreatedAt) || !receipt.AgentCreatedAt.Equal(expected.AgentCreatedAt) {
			return AgentClientRecordResponse{}, false
		}
	}
	for _, agent := range committed.Agents {
		if agent.AgentID != response.Agent.AgentID || agent.OwnerPrincipalID != principal.PrincipalID || agent.WorkspaceID != s.workspaceScope(principal) || !agent.CreatedAt.Equal(response.Agent.CreatedAt) {
			continue
		}
		s.publishOnboardingAgentCreate(committed, agent.AgentID)
		s.version = committed.Version
		// Keep the reload fence: unrelated committed parts were not published.
		return AgentClientRecordResponse{SchemaVersion: SchemaVersion, Agent: s.agentForPrincipal(agent, principal)}, true
	}
	return AgentClientRecordResponse{}, false
}
