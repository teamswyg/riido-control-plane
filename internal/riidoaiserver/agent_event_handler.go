package riidoaiserver

import (
	"net/http"
	"time"
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
	receivedAt := time.Now()
	response, err := s.assignment.RecordAgentEvent(r.Context(), agentID, req)
	if err != nil {
		status, message := agentEventErrorResponse(err)
		logAgentEventRejected(agentID, req, err, status)
		writeError(w, status, message)
		return
	}
	committedAt := time.Now()
	if recorder, ok := s.aiAgent.(AIAgentAssignmentEventRecorder); ok {
		if err := recorder.RecordAIAgentAssignmentEvent(r.Context(), agentID, req, response.Event); err != nil {
			writeError(w, http.StatusServiceUnavailable, "assignment event committed but client projection is pending; retry the same event")
			return
		}
	}
	logAgentResponseDelivery(req, response, receivedAt, committedAt, time.Now())
	writeJSON(w, http.StatusOK, response)
}
