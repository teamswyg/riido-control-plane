package riidoaiserver

import "strings"

// forgetOnboardingStateForAgentLocked drops the onboarding provenance, the one-time
// completion record, and the onboarding completion receipts of a deleted agent. The
// next create with the same owner and runtime reuses the agent ID, and a new agent
// must not inherit any of them.
func (s *DevelopmentAIAgentClientStore) forgetOnboardingStateForAgentLocked(agentID string) {
	delete(s.onboardingCreated, agentID)
	delete(s.onboardingCompletions, agentID)
	delete(s.onboardingCompletionThreads, agentID)
	for key, receipt := range s.taskRequestReceipts {
		if receipt.AgentID == agentID && strings.HasPrefix(key, onboardingCompletionReceiptKeyPrefix) {
			delete(s.taskRequestReceipts, key)
		}
	}
}
