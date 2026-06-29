package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAgentHeartbeat(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	if _, ok := s.authorizeRequest(w, r, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionHeartbeat, AgentID: agentID}); !ok {
		return
	}
	var req AgentHeartbeatRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var response AgentHeartbeatResponse
	var events []TaskEvent
	var err error
	if heartbeatStore, ok := s.assignment.(AssignmentHeartbeatEventStore); ok {
		response, events, err = heartbeatStore.HeartbeatAgentWithEvents(r.Context(), agentID, req)
	} else {
		response, err = s.assignment.HeartbeatAgent(r.Context(), agentID, req)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if recorder, ok := s.aiAgent.(AIAgentAssignmentEventRecorder); ok {
		for _, event := range events {
			_ = recorder.RecordAIAgentAssignmentEvent(r.Context(), agentID, AgentEventRequest{
				AssignmentID: event.AssignmentID,
				TaskID:       event.TaskID,
				DaemonID:     req.DaemonID,
				DeviceID:     req.DeviceID,
				RuntimeID:    req.RuntimeID,
				State:        event.State,
				EventType:    event.Type,
				Message:      event.Message,
				Metadata:     event.Metadata,
			}, event)
		}
	}
	writeJSON(w, http.StatusOK, response)
}
