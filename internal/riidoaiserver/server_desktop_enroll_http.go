package riidoaiserver

import "net/http"

func (s Server) handleDesktopDeviceEnroll(w http.ResponseWriter, r *http.Request, workspaceID string) {
	if s.devices == nil {
		writeError(w, http.StatusServiceUnavailable, "device credential store is not configured")
		return
	}
	r = requestWithAIAgentWorkspaceIDAndPath(r, workspaceID, r.URL.Path)
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate})
	if !ok {
		return
	}
	var req EnrollDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.devices.EnrollDeviceCredential(r.Context(), principal, workspaceID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, response)
}
