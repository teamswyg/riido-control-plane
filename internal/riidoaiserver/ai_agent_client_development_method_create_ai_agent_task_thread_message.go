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
	if active, ok := s.activeTaskThreadLocked(taskID); ok && active.ThreadID != threadID {
		return AIAgentTaskActionResponse{}, ErrAIAgentTaskThreadConflict
	}
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AssignmentID:    thread.AssignmentID,
		AgentID:         agent.AgentID,
		ThreadID:        threadID,
		RunID:           thread.RunID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         "agent work continued from task thread message",
	}
	threadWasActive := taskThreadHasActiveStream(thread)
	if !threadWasActive {
		response.RunID = "run-dev-message-" + taskID + "-" + threadID
		response.AssignmentID = strings.TrimSpace(req.AssignmentID)
	}
	if !threadWasActive && (agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued) {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = "agent is busy; task thread message was queued"
	}
	response = actionResponseWithActiveStream(response, principal.WorkspaceID)
	s.upsertTaskThreadMessageFromActionLocked(response, req.SourceMessageID)
	s.appendAgentTaskActionEvent(response)
	return response, nil
}
