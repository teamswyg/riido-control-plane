package riidoaiserver

import (
	"context"
	"errors"
)

func (s Server) createThreadMessageToolApprovalDecision(
	ctx context.Context,
	principal AuthorizationResult,
	taskID string,
	threadID string,
	req CreateAIAgentTaskThreadMessageRequest,
) (AIAgentTaskActionResponse, bool, error) {
	if !threadMessageApprovesToolApproval(req.Body) {
		return AIAgentTaskActionResponse{}, false, nil
	}
	store, ok := s.assignment.(AssignmentToolApprovalStore)
	if !ok {
		return AIAgentTaskActionResponse{}, false, nil
	}
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return AIAgentTaskActionResponse{}, true, err
	}
	thread, err := selectAIAgentThreadForFollowup(threads, threadID)
	if err != nil {
		return AIAgentTaskActionResponse{}, true, err
	}
	approval, ok, err := pendingThreadToolApproval(ctx, store, taskID, thread)
	if err != nil {
		return AIAgentTaskActionResponse{}, true, err
	}
	if !ok {
		if threadAcceptsApprovalLikeFollowup(thread) {
			return AIAgentTaskActionResponse{}, false, nil
		}
		req.AssignmentID = thread.AssignmentID
		req.toolApproval = true
		req.toolApprovalWithoutPending = true
		response, err := s.aiAgent.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, thread.ThreadID, req)
		return response, true, err
	}
	decision := ToolApprovalDecision{
		ApprovalID:   approval.ApprovalID,
		AssignmentID: approval.AssignmentID,
		Decision:     ApprovalDecisionApprove,
		DecidedBy:    principal.PrincipalID,
		Reason:       "approved from AI agent thread message",
	}
	result, saved, err := store.DecideToolApproval(ctx, taskID, decision)
	if err != nil {
		return AIAgentTaskActionResponse{}, true, err
	}
	if result.Status != ApprovalApproved || saved == nil {
		return AIAgentTaskActionResponse{}, true, errors.New("tool approval was not approved")
	}
	req.AssignmentID = approval.AssignmentID
	req.toolApproval = true
	response, err := s.aiAgent.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, thread.ThreadID, req)
	return response, true, err
}

func threadAcceptsApprovalLikeFollowup(thread AIAgentTaskThreadRecord) bool {
	return thread.WorkStatus == AgentWorkStatusWaitingForUser
}
