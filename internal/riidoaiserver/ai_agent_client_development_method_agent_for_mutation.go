package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) agentForMutation(principal AuthorizationResult, agentID string) (AgentClientRecord, bool) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentClientRecord{}, false
	}
	agent, ok := s.agents[agentID]
	if !ok || !s.aiAgentMutableBy(principal, agent) {
		return AgentClientRecord{}, false
	}
	agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
	s.agents[agent.AgentID] = agent
	return agent, true
}
