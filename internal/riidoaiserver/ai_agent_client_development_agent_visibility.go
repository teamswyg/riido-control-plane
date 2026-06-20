package riidoaiserver

import "slices"

func (s *DevelopmentAIAgentClientStore) aiAgentVisibleTo(principal AuthorizationResult, agent AgentClientRecord) bool {
	if s.agentWorkspaceID(agent) != s.workspaceScope(principal) {
		return false
	}
	if aiAgentIsAdmin(principal) || agent.OwnerPrincipalID == principal.PrincipalID {
		return true
	}
	return agent.Visibility == AgentVisibilityPublic
}

func (s *DevelopmentAIAgentClientStore) aiAgentMutableBy(principal AuthorizationResult, agent AgentClientRecord) bool {
	if s.agentWorkspaceID(agent) != s.workspaceScope(principal) {
		return false
	}
	return aiAgentIsAdmin(principal) || agent.OwnerPrincipalID == principal.PrincipalID
}

func aiAgentIsAdmin(principal AuthorizationResult) bool {
	return slices.Contains(principal.Roles, AgentCatalogRoleAdmin)
}

func normalizeClientKind(kind ClientKind) ClientKind {
	switch kind {
	case ClientKindDesktopWebview:
		return ClientKindDesktopWebview
	default:
		return ClientKindWeb
	}
}

func editabilityForAssignedTasks(count int) AgentEditability {
	if count > 0 {
		return AgentEditabilityBlockedAssignedTasks
	}
	return AgentEditabilityEditable
}
