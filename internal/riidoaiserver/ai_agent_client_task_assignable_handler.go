package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientTaskAssignableAgents(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID})
	if !ok {
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskAssignableAgents(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}
