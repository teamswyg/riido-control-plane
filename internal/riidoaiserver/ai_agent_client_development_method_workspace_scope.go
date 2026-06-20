package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) workspaceScope(principal AuthorizationResult) string {
	if workspaceID := strings.TrimSpace(principal.WorkspaceID); workspaceID != "" {
		return workspaceID
	}
	if s != nil && strings.TrimSpace(s.workspaceID) != "" {
		return strings.TrimSpace(s.workspaceID)
	}
	return defaultAIAgentClientWorkspaceID
}
