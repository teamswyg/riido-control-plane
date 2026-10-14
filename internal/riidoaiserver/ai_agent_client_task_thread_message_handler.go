package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientCreateTaskThreadMessage(w http.ResponseWriter, r *http.Request, taskID, threadID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionCreate, TaskID: taskID})
	if !ok {
		return
	}
	bearerToken, _ := requestToken(r)
	var req CreateAIAgentTaskThreadMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, handled, err := s.createThreadMessageToolApprovalDecision(r.Context(), principal, taskID, threadID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	if handled {
		writeJSON(w, http.StatusAccepted, response)
		return
	}
	assignmentReq, err := s.assignRequestFromAIAgentTaskThreadMessage(r.Context(), principal, bearerToken, taskID, threadID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	assignment, err := s.assignment.AssignTask(r.Context(), taskID, assignmentReq)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	req.AssignmentID = assignment.ID
	req.durableState = assignment.State
	response, err = s.aiAgent.CreateAIAgentTaskThreadMessage(r.Context(), principal, taskID, threadID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
