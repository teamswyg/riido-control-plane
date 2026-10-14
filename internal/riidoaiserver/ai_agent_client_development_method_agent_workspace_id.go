package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) agentWorkspaceID(agent AgentClientRecord) string {
	if workspaceID := strings.TrimSpace(agent.WorkspaceID); workspaceID != "" {
		return workspaceID
	}
	return s.workspaceScope(AuthorizationResult{})
}
