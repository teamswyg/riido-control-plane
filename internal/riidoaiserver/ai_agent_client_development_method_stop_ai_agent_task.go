package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) StopAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req StopAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.agentForTaskStopLocked(principal, taskID, req.AgentID, req.AssignmentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	workStatus, assignmentState, completed := stopReadModelProjection(req.durableState)
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AgentID:         agent.AgentID,
		RunID:           "run-dev-stop-" + taskID,
		WorkStatus:      workStatus,
		AssignmentState: assignmentState,
		CommentKind:     AgentTaskCommentStoppedByUserRequest,
		Message:         clientMessageTaskStopped,
	}
	if thread, ok := s.taskThreadForStopTargetLocked(taskID, agent.AgentID, req.AssignmentID); ok {
		response.ThreadID = thread.ThreadID
		response.AssignmentID = thread.AssignmentID
		response.RunID = thread.RunID
	} else if req.AssignmentID != "" {
		return AIAgentTaskActionResponse{}, errors.New("assignment_id does not belong to task agent")
	} else {
		response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	}
	if req.AssignmentID != "" {
		s.markTaskAgentAssignmentThreadStopProjectionLocked(taskID, agent.AgentID, req.AssignmentID, response, completed)
	} else {
		s.markTaskAgentThreadsStopProjectionLocked(taskID, agent.AgentID, response, completed)
	}
	response = actionResponseWithActiveStream(response, principal.WorkspaceID)
	s.upsertTaskThreadFromActionLocked(response, "")
	agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
	s.agents[agent.AgentID] = agent
	s.appendAgentTaskActionEvent(response)
	return response, nil
}
