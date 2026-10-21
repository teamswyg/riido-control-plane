package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s Server) readTaskThreadMessageContext(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskThreadRecord, AIAgentTaskThreadHistoryCollectionResponse, error) {
	taskID, threadID = strings.TrimSpace(taskID), strings.TrimSpace(threadID)
	if taskID == "" || threadID == "" || strings.TrimSpace(req.Body) == "" {
		return AIAgentTaskThreadRecord{}, AIAgentTaskThreadHistoryCollectionResponse{}, errors.New("task_id, thread_id and body are required")
	}
	if err := s.reconcileAIAgentTaskThreadProjections(ctx, principal, taskID); err != nil {
		return AIAgentTaskThreadRecord{}, AIAgentTaskThreadHistoryCollectionResponse{}, err
	}
	history, err := s.aiAgent.ListAIAgentTaskThreadHistory(ctx, principal, taskID)
	if err != nil {
		return AIAgentTaskThreadRecord{}, history, err
	}
	selected, err := selectAIAgentHistoryThreadForFollowup(history, taskID, threadID)
	return selected, history, err
}

func (s Server) assignRequestFromAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, bearerToken, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest, selectedThread AIAgentTaskThreadRecord, history AIAgentTaskThreadHistoryCollectionResponse) (AssignRequest, error) {
	composed, err := s.assignRequestFromAIAgentClientTaskResult(ctx, principal, bearerToken, taskID, AssignAIAgentTaskRequest{
		AgentID:          selectedThread.AgentID,
		RequestID:        threadMessageRequestID(threadID, req.RequestID),
		DocumentSnapshot: req.DocumentSnapshot,
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
