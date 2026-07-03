package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) handleAIAgentClientDeviceRoutes(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	deviceID, suffix, ok := splitAIAgentClientDevicePath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "":
		writeMethodNotAllowed(w)
	case suffix == "daemon" && r.Method == http.MethodGet:
		s.handleAIAgentClientDeviceDaemon(w, r, deviceID)
	case suffix == "daemons" && r.Method == http.MethodGet:
		s.handleAIAgentClientDeviceDaemons(w, r, deviceID)
	case strings.HasPrefix(suffix, "daemon/") && r.Method == http.MethodPost:
		s.handleAIAgentClientDeviceDaemonControl(w, r, deviceID, strings.TrimPrefix(suffix, "daemon/"))
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s Server) handleAIAgentClientDeviceDaemon(w http.ResponseWriter, r *http.Request, deviceID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead, DeviceID: deviceID})
	if !ok {
		return
	}
	response, err := s.aiAgent.GetAIAgentDeviceDaemon(r.Context(), principal, deviceID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientDeviceDaemonControl(w http.ResponseWriter, r *http.Request, deviceID, actionValue string) {
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
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceControl, DeviceID: deviceID})
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
	response, err := s.aiAgent.ControlAIAgentDeviceDaemon(r.Context(), principal, deviceID, action, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
