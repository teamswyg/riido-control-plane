package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) SubmitAIAgentTaskComment(ctx context.Context, principal AuthorizationResult, taskID string, req SubmitAIAgentTaskCommentRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.Body = strings.TrimSpace(req.Body)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AIAgentTaskActionResponse{}, errors.New("agent_id is required")
	}
	if req.Body == "" {
		return AIAgentTaskActionResponse{}, errors.New("body is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.visibleAgent(principal, req.AgentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AgentID:         agent.AgentID,
		AgentSnapshot:   s.agentSnapshotFromAgent(agent, time.Now().UTC()),
		RunID:           "run-dev-comment-" + taskID,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentRuntimeProgress,
		Message:         clientMessageTaskRunning,
	}
	response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	if agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = clientMessageAgentBusyQueued
	}
	response = actionResponseWithActiveStream(response, principal.WorkspaceID)
	s.upsertTaskThreadFromActionLocked(response, req.SourceCommentID)
	s.appendAgentTaskActionEvent(response)
	return response, nil
}
