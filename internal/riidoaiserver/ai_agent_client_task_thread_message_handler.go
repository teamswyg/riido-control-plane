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
	if len(req.RequestID) > 256 {
		writeError(w, http.StatusBadRequest, "request_id maximum is 256 bytes")
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
	prepared, err := s.prepareThreadMessage(r.Context(), principal, bearerToken, taskID, threadID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	if prepared.Replay != nil {
		writeJSON(w, http.StatusAccepted, *prepared.Replay)
		return
	}
	creationContext, creationReceipt := withTaskAgentCreationReceipt(r.Context())
	assignment, err := s.assignThreadMessage(creationContext, taskID, prepared.Request, prepared.StartsNewExecution)
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
	// The thread record maps this assignment to its source message.
	s.captureAssignmentContext(r.Context(), assignment, prepared.TaskContext, assignmentContextTrigger{Kind: AssignmentContextTriggerThreadMessage, ID: threadID})
	writeJSON(w, http.StatusAccepted, response)
}
