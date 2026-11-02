package riidoaiserver

import "net/http"

func parseAssignmentContextRequest(w http.ResponseWriter, r *http.Request) (AgentAssignmentContextRequest, assignmentContextCache, bool) {
	var req AgentAssignmentContextRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return req, assignmentContextCache{}, false
	}
	if !normalizeAssignmentContextRequest(&req) {
		writeError(w, http.StatusBadRequest, "assignment_id, lease_token and a supported view are required")
		return req, assignmentContextCache{}, false
	}
	cache, err := assignmentContextCacheRequest(r.Header, req.View)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid assignment context cache digest")
		return req, assignmentContextCache{}, false
	}
	return req, cache, true
}
