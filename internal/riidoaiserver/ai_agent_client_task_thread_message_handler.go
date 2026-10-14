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
	release, err := s.taskAgentExecutionGate.beginPublication(r.Context(), taskID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	defer release()
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
	creationContext, creationReceipt := withTaskAgentCreationReceipt(r.Context())
	assignment, err := s.assignment.AssignTask(creationContext, taskID, assignmentReq)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	req.AssignmentID = assignment.ID
	if owner, ok := s.assignment.(interface{ taskAgentGenerationSupported() bool }); ok && owner.taskAgentGenerationSupported() && assignmentSequence(assignment.ID) == 0 {
		req.membershipPublicationGeneration = taskAgentAssignmentGeneration(assignment)
	}
	req.durableState = assignment.State
	response, err = s.aiAgent.CreateAIAgentTaskThreadMessage(r.Context(), principal, taskID, threadID, req)
	if err != nil {
		s.recoverTaskAgentPublication(r.Context(), assignment, creationReceipt)
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
