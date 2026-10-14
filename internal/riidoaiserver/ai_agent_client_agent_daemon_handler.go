package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientAgentDaemon(w http.ResponseWriter, r *http.Request, agentID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead, AgentID: agentID})
	if !ok {
		return
	}
	response, err := s.aiAgent.GetAIAgentDaemon(r.Context(), principal, agentID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientAgentDaemonControl(w http.ResponseWriter, r *http.Request, agentID, actionValue string) {
	var action DaemonControlAction
	switch actionValue {
	case string(DaemonControlActionStart):
		action = DaemonControlActionStart
	case string(DaemonControlActionRestart):
		action = DaemonControlActionRestart
	case string(DaemonControlActionStop):
		action = DaemonControlActionStop
	default:
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceControl, AgentID: agentID})
	if !ok {
		return
	}
	var req ControlDeviceDaemonRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	response, err := s.aiAgent.ControlAIAgentDaemon(r.Context(), principal, agentID, action, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
