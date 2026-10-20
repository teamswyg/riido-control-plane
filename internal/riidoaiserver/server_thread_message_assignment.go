package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s Server) assignRequestFromAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, bearerToken, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AssignRequest, error) {
	taskID = strings.TrimSpace(taskID)
	threadID = strings.TrimSpace(threadID)
	req.Body = strings.TrimSpace(req.Body)
	if taskID == "" {
		return AssignRequest{}, errors.New("task_id is required")
	}
	if threadID == "" {
		return AssignRequest{}, errors.New("thread_id is required")
	}
	if req.Body == "" {
		return AssignRequest{}, errors.New("body is required")
	}
	if err := s.reconcileAIAgentTaskThreadProjections(ctx, principal, taskID); err != nil {
		return AssignRequest{}, err
	}
	history, err := s.aiAgent.ListAIAgentTaskThreadHistory(ctx, principal, taskID)
	if err != nil {
		return AssignRequest{}, err
	}
	selectedThread, err := selectAIAgentHistoryThreadForFollowup(history, taskID, threadID)
	if err != nil {
		return AssignRequest{}, err
	}
	composed, err := s.assignRequestFromAIAgentClientTaskResult(ctx, principal, bearerToken, taskID, AssignAIAgentTaskRequest{
		AgentID: selectedThread.AgentID,
	})
	if err != nil {
		return AssignRequest{}, err
	}
	assignmentReq := composed.Request
	if err := s.applyThreadNativeResume(ctx, principal, selectedThread, &assignmentReq, composed.AgentOwnerPrincipalID); err != nil {
		return AssignRequest{}, err
	}
	assignmentReq.Prompt = appendAIAgentTaskThreadMessagePrompt(assignmentReq.Prompt, selectedThread, req, history.Threads...)
	return assignmentReq, nil
}

func selectAIAgentThreadForFollowup(threads AIAgentTaskThreadCollectionResponse, threadID string) (AIAgentTaskThreadRecord, error) {
	for _, thread := range threads.Threads {
		if thread.ThreadID == threadID {
			return thread, nil
		}
	}
	return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
}
