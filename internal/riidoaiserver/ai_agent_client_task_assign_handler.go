package riidoaiserver

import (
	"net/http"
)

func (s Server) handleAIAgentClientAssignTask(w http.ResponseWriter, r *http.Request, taskID string) {
	if s.assignment == nil {
		writeError(w, http.StatusServiceUnavailable, "assignment store is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionAssign, TaskID: taskID})
	if !ok {
		return
	}
	bearerToken, _ := requestToken(r)
	var req AssignAIAgentTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	composedReq, err := s.assignRequestFromAIAgentClientTaskResult(r.Context(), principal, bearerToken, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	if composedReq.IntentGateRequired {
		if _, _, err := s.cancelAIAgentAssignmentBeforeAction(r.Context(), principal, taskID, "", "", "intent_gate_required"); err != nil {
			writeAIAgentClientError(w, err)
			return
		}
		req.intentGateRequired = true
		response, err := s.aiAgent.AssignAIAgentTask(r.Context(), principal, taskID, req)
		if err != nil {
			writeAIAgentClientError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, response)
		return
	}
	assignmentReq := composedReq.Request
	assignment, err := s.assignment.AssignTaskReplacement(r.Context(), taskID, assignmentReq)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	req.AssignmentID, req.durableState = assignment.ID, assignmentClientResponseDurableState(assignment)
	response, err := s.aiAgent.AssignAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
