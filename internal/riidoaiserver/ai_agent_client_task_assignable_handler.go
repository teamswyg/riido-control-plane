package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientTaskAssignableAgents(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, ""); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskAssignableAgents(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Agents = s.runtimeBoundAssignableAgents(response.Agents)
	writeJSON(w, http.StatusOK, response)
}
