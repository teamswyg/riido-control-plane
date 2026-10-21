package riidoaiserver

import "time"

func (s *PersistentAIAgentClientStore) requireOnboardingCreateReload() {
	s.snapshotReloadMu.Lock()
	s.snapshotLastReloadAt = time.Time{}
	s.snapshotReloadRequired = true
	s.snapshotReloadMu.Unlock()
}

func (s *PersistentAIAgentClientStore) publishOnboardingAgentCreate(snapshot AIAgentClientSnapshot, agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, agent := range snapshot.Agents {
		if agent.AgentID == agentID {
			s.agents[agentID] = agent
			break
		}
	}
	s.markOnboardingCreatedAgentLocked(agentID)
	s.onboardingAgentCreates = cloneOnboardingAgentCreates(snapshot.OnboardingAgentCreates)
}
