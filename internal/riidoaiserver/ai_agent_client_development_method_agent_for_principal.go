package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) agentForPrincipal(agent AgentClientRecord, principal AuthorizationResult) AgentClientRecord {
	agent.ReferenceFolders = agent.ReferenceFolders.Clone()
	agent.IsOwnedByViewer = agent.OwnerPrincipalID == principal.PrincipalID
	agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
	if strings.TrimSpace(principal.WorkspaceID) == "" {
		agent.WorkspaceID = ""
	}
	return agent
}
