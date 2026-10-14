package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) taskThreadWorkspaceIDLocked(thread AIAgentTaskThreadRecord) string {
	if agent, ok := s.agents[thread.AgentID]; ok {
		return s.agentWorkspaceID(agent)
	}
	if thread.AgentSnapshot != nil {
		return strings.TrimSpace(thread.AgentSnapshot.WorkspaceID)
	}
	return ""
}
