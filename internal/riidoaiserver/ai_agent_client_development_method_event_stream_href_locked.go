package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) eventStreamHrefLocked(workspaceID string) string {
	workspaceID = strings.TrimSpace(workspaceID)
	if s.eventStreamHrefs == nil {
		s.eventStreamHrefs = map[string]string{}
	}
	if href, ok := s.eventStreamHrefs[workspaceID]; ok {
		return href
	}
	href := aiAgentClientEventStreamHref(workspaceID)
	s.eventStreamHrefs[workspaceID] = href
	return href
}
