package riidoaiserver

import "net/http"

func (s Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead}); !ok {
		return
	}
	snapshot, err := s.assignment.Metrics(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	snapshot = s.config.HTTPTransactions.ApplyToMetricsSnapshot(snapshot)
	writeJSON(w, http.StatusOK, snapshot)
}
