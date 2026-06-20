package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) agentForPrincipal(agent AgentClientRecord, principal AuthorizationResult) AgentClientRecord {
	agent.IsOwnedByViewer = agent.OwnerPrincipalID == principal.PrincipalID
	agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
	if strings.TrimSpace(principal.WorkspaceID) == "" {
		agent.WorkspaceID = ""
	}
	return agent
}
