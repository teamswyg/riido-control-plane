package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) visibleAgent(principal AuthorizationResult, agentID string) (AgentClientRecord, bool) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return AgentClientRecord{}, false
	}
	agent, ok := s.agents[agentID]
	if !ok || !s.aiAgentVisibleTo(principal, agent) {
		return AgentClientRecord{}, false
	}
	agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
	s.agents[agent.AgentID] = agent
	return s.agentForPrincipal(agent, principal), true
}
