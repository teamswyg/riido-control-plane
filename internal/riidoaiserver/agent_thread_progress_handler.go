package riidoaiserver

import "net/http"

func (s Server) handleAgentThreadProgress(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, AgentID: agentID}); !ok {
		return
	}
	var req AgentThreadProgressBatchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := normalizeAgentThreadProgressRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.recordAgentThreadProgressEvents(r.Context(), agentID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if recorder, ok := s.aiAgent.(AIAgentThreadProgressRecorder); ok {
		response, err := recorder.RecordAIAgentThreadProgress(r.Context(), agentID, req)
		if err != nil {
			writeAIAgentClientError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, response)
		return
	}
	writeJSON(w, http.StatusAccepted, fallbackAgentThreadProgressResponse(agentID, req))
}
