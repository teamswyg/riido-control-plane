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
	release, err := s.taskAgentExecutionGate.beginPublication(r.Context(), taskID)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	defer release()
	if err := s.reconcileAIAgentTaskThreadProjections(r.Context(), principal, taskID); err != nil {
		writeThreadProjectionError(w, r, err)
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
	creationContext, creationReceipt := withTaskAgentCreationReceipt(r.Context())
	assignment, err := s.assignment.AssignTaskReplacement(creationContext, taskID, assignmentReq)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	req.AssignmentID, req.durableState = assignment.ID, assignmentClientResponseDurableState(assignment)
	if owner, ok := s.assignment.(interface{ taskAgentGenerationSupported() bool }); ok && owner.taskAgentGenerationSupported() && assignmentSequence(assignment.ID) == 0 {
		req.membershipPublicationGeneration = taskAgentAssignmentGeneration(assignment)
	}
	response, err := s.aiAgent.AssignAIAgentTask(r.Context(), principal, taskID, req)
	if err != nil {
		s.recoverTaskAgentPublication(r.Context(), assignment, creationReceipt)
		writeAIAgentClientError(w, err)
		return
	}
	s.captureAssignmentContext(r.Context(), assignment, composedReq.TaskContext, assignmentContextTrigger{Kind: AssignmentContextTriggerAssign, ID: req.RequestID, RequesterAdmin: aiAgentIsAdmin(principal)})
	writeJSON(w, http.StatusAccepted, response)
}
