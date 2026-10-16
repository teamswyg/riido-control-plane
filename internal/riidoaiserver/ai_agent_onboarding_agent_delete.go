package riidoaiserver

// forgetOnboardingStateForAgentLocked drops the onboarding provenance and the one-time
// completion record of a deleted agent. The next create with the same owner and runtime
// reuses the agent ID, and a new agent must not inherit either record.
func (s *DevelopmentAIAgentClientStore) forgetOnboardingStateForAgentLocked(agentID string) {
	delete(s.onboardingCreated, agentID)
	delete(s.onboardingCompletions, agentID)
	delete(s.onboardingCompletionThreads, agentID)
}
