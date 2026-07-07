package riidoaiserver

import (
	"net/http"
)

func (s Server) toolApprovalStore(w http.ResponseWriter) (AssignmentToolApprovalStore, bool) {
	store, ok := s.assignment.(AssignmentToolApprovalStore)
	if s.assignment == nil || !ok {
		writeError(w, http.StatusServiceUnavailable, "tool approval store is not configured")
		return nil, false
	}
	return store, true
}
