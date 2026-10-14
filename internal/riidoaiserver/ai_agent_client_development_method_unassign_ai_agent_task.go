package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) UnassignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AIAgentTaskActionResponse{}, errors.New("agent_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.visibleAgent(principal, req.AgentID)
	if !ok {
		if thread, found := s.taskThreadForUnassignTargetLocked(taskID, req.AgentID, req.AssignmentID); found {
			agent, ok = s.agentFromTaskThreadLocked(principal, thread)
		}
		if !ok {
			return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
		}
	}
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AgentID:         agent.AgentID,
		RunID:           "run-dev-unassign-" + taskID,
		WorkStatus:      AgentWorkStatusIdle,
		AssignmentState: AgentAssignmentStateStopped,
		CommentKind:     AgentTaskCommentStoppedByUserRequest,
		Message:         clientMessageTaskStopped,
	}
	if thread, ok := s.taskThreadForUnassignTargetLocked(taskID, agent.AgentID, req.AssignmentID); ok {
		response.ThreadID = thread.ThreadID
		response.AssignmentID = thread.AssignmentID
		response.RunID = thread.RunID
	} else if req.AssignmentID != "" {
		return AIAgentTaskActionResponse{}, errors.New("assignment_id does not belong to task agent")
	} else {
		response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	}
	if req.AssignmentID != "" {
		s.markTaskAgentAssignmentThreadStoppedLocked(taskID, agent.AgentID, req.AssignmentID, AgentTaskCommentStoppedByUserRequest, response.Message)
	} else {
		s.markTaskAgentThreadsStoppedLocked(taskID, agent.AgentID, AgentTaskCommentStoppedByUserRequest, response.Message)
	}
	s.upsertTaskThreadFromActionLocked(response, "")
	if _, ok := s.agents[agent.AgentID]; ok {
		agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
		s.agents[agent.AgentID] = agent
	}
	s.appendAgentTaskActionEvent(response)
	return response, nil
}
