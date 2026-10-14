package riidoaiserver

import "net/http"

func (s Server) handleAIAgentClientWorkspaceAssignedAgentProfiles(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, ""); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.ListWorkspaceAssignedAgentProfiles(r.Context(), principal)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}
