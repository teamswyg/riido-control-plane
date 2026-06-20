package riidoaiserver

import (
	"sort"
)

func (s *DevelopmentAIAgentClientStore) visibleAgents(principal AuthorizationResult) []AgentClientRecord {
	agents := make([]AgentClientRecord, 0, len(s.agents))
	for _, agent := range s.agents {
		if !s.aiAgentVisibleTo(principal, agent) {
			continue
		}
		agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
		s.agents[agent.AgentID] = agent
		agents = append(agents, s.agentForPrincipal(agent, principal))
	}
	sort.SliceStable(agents, func(i, j int) bool {
		if agents[i].IsOwnedByViewer != agents[j].IsOwnedByViewer {
			return agents[i].IsOwnedByViewer
		}
		if agents[i].Name != agents[j].Name {
			return agents[i].Name < agents[j].Name
		}
		return agents[i].AgentID < agents[j].AgentID
	})
	return agents
}
