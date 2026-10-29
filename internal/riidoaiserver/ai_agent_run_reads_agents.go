package riidoaiserver

import (
	"net/http"
	"slices"
	"strings"
)

// addVisibleAgentReadSources shows agent-directory reads only for agents the
// viewer can see now. The directory is listed only when such reads exist, and
// a listing failure fails closed instead of exposing or silently dropping them.
func (s Server) addVisibleAgentReadSources(w http.ResponseWriter, r *http.Request, viewer AuthorizationResult, taskID string, reads []AssignmentContextRead, visible map[runReadSourceKey]runReadMetadata) bool {
	if !slices.ContainsFunc(reads, func(read AssignmentContextRead) bool { return read.SourceKind == "agent" }) {
		return true
	}
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "run reads unavailable")
		return false
	}
	list, err := s.aiAgent.ListAIAgentTaskAssignableAgents(r.Context(), viewer, taskID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "run reads unavailable")
		return false
	}
	for _, agent := range list.Agents {
		entry := AgentDirectoryEntry{AgentID: agent.AgentID, Name: strings.TrimSpace(agent.Name), Description: strings.TrimSpace(agent.Description), UpdatedAt: agent.UpdatedAt}
		visible[runReadSourceKey{"agent", agent.AgentID, "profile"}] = runReadMetadata{title: entry.Name, source: agentDirectorySource(entry), comparable: true}
	}
	return true
}
