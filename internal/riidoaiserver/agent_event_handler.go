package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAgentEvent(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, AgentID: agentID}); !ok {
		return
	}
	var req AgentEventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.assignment.RecordAgentEvent(r.Context(), agentID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if recorder, ok := s.aiAgent.(AIAgentAssignmentEventRecorder); ok {
		_ = recorder.RecordAIAgentAssignmentEvent(r.Context(), agentID, req, response.Event)
	}
	writeJSON(w, http.StatusOK, response)
}
