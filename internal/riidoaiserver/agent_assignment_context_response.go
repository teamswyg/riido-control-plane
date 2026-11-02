package riidoaiserver

import "net/http"

// writeAssignmentContextRepresentation runs only after lookup, quota and audit.
// The cache candidate is a representation hint, never read authority.
func writeAssignmentContextRepresentation(w http.ResponseWriter, cache assignmentContextCache, agentID string, req AgentAssignmentContextRequest, response AgentAssignmentContextResponse) {
	if err := applyAssignmentContextCache(cache, agentID, req, &response); err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context representation unavailable")
		return
	}
	writeJSON(w, http.StatusOK, response)
}
