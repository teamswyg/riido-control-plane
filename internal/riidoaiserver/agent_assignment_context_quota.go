package riidoaiserver

import "net/http"

// Charge after authentication and active-lease validation, before any live or
// snapshot read. Quota failures must never fall through to data access.
func (s Server) consumeAssignmentContextRead(w http.ResponseWriter, r *http.Request, assignmentID string) bool {
	allowed, err := s.config.AssignmentContext.ConsumeAssignmentContextRead(r.Context(), assignmentID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assignment context quota unavailable")
		return false
	}
	if !allowed {
		writeError(w, http.StatusTooManyRequests, "assignment context read limit exceeded; start a new execution")
		return false
	}
	return true
}
