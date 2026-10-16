package riidoaiserver

// Deletion is an incarnation-wide intent, not only removal of keys known to a
// cached reader. A completion that committed since deletion loaded its baseline
// must also be cleared, while the captured historical task comment remains.
func forgetDeletedOnboardingIncarnations(baseline, requested AIAgentClientSnapshot, committed *AIAgentClientSnapshot) {
	retained := onboardingCreatedAgentsFromSnapshot(requested.OnboardingCreatedIDs)
	for _, agentID := range baseline.OnboardingCreatedIDs {
		if retained[agentID] {
			continue
		}
		delete(committed.OnboardingCompletions, agentID)
		delete(committed.OnboardingCompletionThreads, agentID)
		for key, receipt := range committed.TaskRequestReceipts {
			if receipt.AgentID == agentID {
				delete(committed.TaskRequestReceipts, key)
			}
		}
	}
}
