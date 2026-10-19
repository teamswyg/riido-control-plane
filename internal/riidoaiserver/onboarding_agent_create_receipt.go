package riidoaiserver

import (
	"maps"
	"time"
)

func cloneOnboardingAgentCreates(in map[string]onboardingAgentCreateReceipt) map[string]onboardingAgentCreateReceipt {
	return maps.Clone(in)
}

// A receipt outlives deletion of its agent. A retry must never resurrect a
// deleted result, and must not depend on the provider still being online.
func (s *DevelopmentAIAgentClientStore) replayOnboardingAgentCreateLocked(principal AuthorizationResult, key, fingerprint string, now time.Time) (AgentClientRecordResponse, bool, error) {
	receipt, found := s.onboardingAgentCreates[key]
	if !found || now.Compare(receipt.CreatedAt.Add(onboardingAgentCreateRetention)) >= 0 {
		return AgentClientRecordResponse{}, false, nil
	}
	if receipt.Fingerprint != fingerprint {
		return AgentClientRecordResponse{}, true, errOnboardingAgentCreateConflict
	}
	agent, found := s.agents[receipt.AgentID]
	if !found || agent.OwnerPrincipalID != principal.PrincipalID || agent.WorkspaceID != s.workspaceScope(principal) || !agent.CreatedAt.Equal(receipt.AgentCreatedAt) {
		return AgentClientRecordResponse{}, true, errOnboardingAgentCreateConflict
	}
	return AgentClientRecordResponse{SchemaVersion: SchemaVersion, Agent: s.agentForPrincipal(agent, principal)}, true, nil
}

func retainOnboardingAgentCreates(in map[string]onboardingAgentCreateReceipt, now time.Time) map[string]onboardingAgentCreateReceipt {
	out := map[string]onboardingAgentCreateReceipt{}
	for k, receipt := range in {
		if now.Before(receipt.CreatedAt.Add(onboardingAgentCreateRetention)) {
			out[k] = receipt
		}
	}
	return out
}
