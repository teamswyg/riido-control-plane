package riidoaiserver

import "net/http"

func (s Server) handleAIAgentClientDeviceDaemons(w http.ResponseWriter, r *http.Request, deviceID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead, DeviceID: deviceID})
	if !ok {
		return
	}
	response, err := s.aiAgent.ListAIAgentDeviceDaemons(r.Context(), principal, deviceID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}
