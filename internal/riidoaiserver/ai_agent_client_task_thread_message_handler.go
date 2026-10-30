package riidoaiserver

import (
	"context"
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
	replies, err := s.prepareThreadReplies(r.Context(), principal, bearerToken, taskID, threadID, req)
	if err != nil {
		writeAIAgentClientError(w, err)
		return
	}
	for index, reply := range replies {
		result, err := s.publishThreadReply(r.Context(), principal, taskID, threadID, reply)
		if err != nil {
			writeAIAgentClientError(w, err)
			return
		}
		if index == 0 {
			response = result
		}
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s Server) publishThreadReply(ctx context.Context, principal AuthorizationResult, taskID, threadID string, reply preparedThreadReply) (AIAgentTaskActionResponse, error) {
	req, prepared := reply.Message, reply.Prepared
	if prepared.Replay != nil {
		return *prepared.Replay, nil
	}
	creationContext, creationReceipt := withTaskAgentCreationReceipt(ctx)
	assignment, err := s.assignThreadMessage(creationContext, taskID, prepared.Request, prepared.StartsNewExecution)
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	req.AssignmentID = assignment.ID
	req.commentInvocation = prepared.CommentInvocation
	if owner, ok := s.assignment.(interface{ taskAgentGenerationSupported() bool }); ok && owner.taskAgentGenerationSupported() && assignmentSequence(assignment.ID) == 0 {
		req.membershipPublicationGeneration = taskAgentAssignmentGeneration(assignment)
	}
	req.durableState = assignment.State
	response, err := s.aiAgent.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, req)
	if err != nil {
		s.recoverTaskAgentPublication(ctx, assignment, creationReceipt)
		return AIAgentTaskActionResponse{}, err
	}
	// A message folded into a running assignment did not get this prompt.
	if prepared.StartsNewExecution || creationReceipt.ownCreation(assignment) {
		s.captureAssignmentContext(ctx, assignment, prepared.TaskContext, assignmentContextTrigger{Kind: AssignmentContextTriggerThreadMessage, ID: threadID, RequesterAdmin: aiAgentIsAdmin(principal)})
	}
	return response, nil
}
