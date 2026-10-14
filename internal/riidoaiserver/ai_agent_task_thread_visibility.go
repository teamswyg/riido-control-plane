package riidoaiserver

func (s *DevelopmentAIAgentClientStore) taskThreadVisibleTo(principal AuthorizationResult, thread AIAgentTaskThreadRecord) bool {
	if agent, ok := s.agents[thread.AgentID]; ok {
		return s.aiAgentVisibleTo(principal, agent)
	}
	snapshot := thread.AgentSnapshot
	if snapshot == nil || snapshot.WorkspaceID != s.workspaceScope(principal) {
		return false
	}
	if aiAgentIsAdmin(principal) || snapshot.OwnerPrincipalID == principal.PrincipalID {
		return true
	}
	return snapshot.Visibility == AgentVisibilityPublic
}
