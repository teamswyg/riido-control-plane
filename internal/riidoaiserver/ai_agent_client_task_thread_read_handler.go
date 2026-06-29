package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientTaskThreads(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskThreads(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientTaskThreadStreamSubscription(w http.ResponseWriter, r *http.Request, taskID string) {
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, TaskID: taskID})
	if !ok {
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	response, err := s.aiAgent.GetAIAgentTaskThreadStreamSubscription(r.Context(), principal, taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s Server) handleAIAgentClientSubmitTaskComment(w http.ResponseWriter, r *http.Request, taskID string) {
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate, TaskID: taskID})
	if !ok {
		return
	}
	var req SubmitAIAgentTaskCommentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.aiAgent.SubmitAIAgentTaskComment(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
