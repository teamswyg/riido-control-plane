package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	threadID = strings.TrimSpace(threadID)
	req.Body = strings.TrimSpace(req.Body)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if threadID == "" {
		return AIAgentTaskActionResponse{}, errors.New("thread_id is required")
	}
	if req.Body == "" {
		return AIAgentTaskActionResponse{}, errors.New("body is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.visibleTaskThreadLocked(principal, taskID, threadID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	agent, ok := s.visibleAgent(principal, thread.AgentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AssignmentID:    thread.AssignmentID,
		AgentID:         agent.AgentID,
		AgentSnapshot:   copyTaskThreadAgentSnapshot(thread.AgentSnapshot),
		ThreadID:        threadID,
		RunID:           thread.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         clientMessageTaskRunning,
	}
	threadWasActive := taskThreadHasActiveStream(thread)
	conversationID := taskThreadConversationID(thread)
	parentThreadID := ""
	if !req.toolApproval && taskThreadMessageStartsNewExecution(thread) {
		parentThreadID = thread.ThreadID
		response.AssignmentID = strings.TrimSpace(req.AssignmentID)
		response.RunID = taskThreadMessageRunID(taskID, response.AssignmentID, len(s.taskThreads[taskID])+1)
		response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	}
	if !threadWasActive && (agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued) {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = clientMessageAgentBusyQueued
	}
	if assignmentStateIsKnown(req.durableState) {
		response.Message = ""
		applyAssignmentStateActionResponse(&response, req.durableState)
	}
	if req.toolApprovalWithoutPending {
		applyToolApprovalWithoutPendingActionResponse(&response)
	}
	response = actionResponseWithActiveStream(response, principal.WorkspaceID)
	s.appendThreadUserMessageLocked(response, principal, req.Body, req.SourceMessageID)
	s.upsertTaskThreadMessageFromActionLocked(response, req.SourceMessageID, conversationID, parentThreadID)
	s.appendAgentTaskActionEvent(response)
	return response, nil
}
