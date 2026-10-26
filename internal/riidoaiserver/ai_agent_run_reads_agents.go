package riidoaiserver

import (
	"net/http"
	"slices"
)

// addVisibleAgentReadSources shows agent-directory reads only for agents the
// viewer can see now. The directory is listed only when such reads exist, and
// a listing failure fails closed instead of exposing or silently dropping them.
func (s Server) addVisibleAgentReadSources(w http.ResponseWriter, r *http.Request, viewer AuthorizationResult, taskID string, reads []AssignmentContextRead, visible map[runReadSourceKey]bool) bool {
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
		visible[runReadSourceKey{"agent", agent.AgentID, "profile"}] = true
	}
	return true
}
