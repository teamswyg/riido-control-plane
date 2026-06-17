package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) reprojectAgentsFromTaskThreadsLocked() {
	for agentID, agent := range s.agents {
		if strings.TrimSpace(agent.AgentID) == "" {
			agent.AgentID = agentID
		}
		s.agents[agentID] = s.projectAgentWorkStatusFromThreadsLocked(agent)
	}
}
