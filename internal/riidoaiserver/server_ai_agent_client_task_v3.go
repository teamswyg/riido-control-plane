package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientTasksV3(w http.ResponseWriter, r *http.Request) {
	if s.aiAgent == nil {
		writeError(w, http.StatusServiceUnavailable, "ai agent client store is not configured")
		return
	}
	taskID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v3/client/ai-agent/tasks/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch {
	case suffix == "threads" && r.Method == http.MethodGet:
		s.handleAIAgentClientTaskThreadsV3(w, r, taskID)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s Server) handleAIAgentClientTaskThreadsV3(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient,
		Action:   AuthorizationActionRead,
		TaskID:   taskID,
	})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskThreadHistory(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}
