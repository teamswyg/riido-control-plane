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
	if err := validateTaskAgentMembershipRemovalObservation(taskID, req.AgentID, req.membershipRemovalFence, s.taskThreads[taskID]); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	agent, ok := s.visibleAgent(principal, req.AgentID)
	if !ok {
		if thread, found := s.taskThreadForUnassignTargetLocked(taskID, req.AgentID, req.AssignmentID); found {
			agent, ok = s.agentFromTaskThreadLocked(principal, thread)
		}
		if !ok && req.AssignmentID == "" {
			if member, found := s.taskAgentMemberships[taskID][req.AgentID]; found && member.RemovedAt == nil {
				agent, ok = s.agentFromTaskThreadLocked(principal, AIAgentTaskThreadRecord{AgentID: member.AgentID, AgentSnapshot: member.AgentSnapshot})
			}
		}
		if !ok {
			return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
		}
	}
	if err := validateMembershipRemovalAssignment(req.AgentID, req.AssignmentID, s.taskThreads[taskID]); err != nil {
		return AIAgentTaskActionResponse{}, err
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
	active := false
	if thread, ok := s.taskThreadForUnassignTargetLocked(taskID, agent.AgentID, req.AssignmentID); ok {
		response.ThreadID = thread.ThreadID
		response.AssignmentID = thread.AssignmentID
		response.RunID = thread.RunID
		active = taskThreadHasActiveStream(thread)
		if !active {
			response.WorkStatus = thread.WorkStatus
			response.AssignmentState = thread.AssignmentState
			response.CommentKind = thread.CommentKind
			response.Message = thread.Message
			response.ResultMessage = thread.ResultMessage
			response.FailureDiagnostics = copyFailureDiagnostics(thread.FailureDiagnostics)
		}
	} else if req.AssignmentID != "" {
		return AIAgentTaskActionResponse{}, errors.New("assignment_id does not belong to task agent")
	} else {
		response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	}
	if req.AssignmentID != "" && req.membershipRemovalFence == nil {
		s.markTaskAgentAssignmentThreadStoppedLocked(taskID, agent.AgentID, req.AssignmentID, AgentTaskCommentStoppedByUserRequest, response.Message)
	} else {
		s.markTaskAgentThreadsStoppedLocked(taskID, agent.AgentID, AgentTaskCommentStoppedByUserRequest, response.Message)
	}
	// Removing membership is not a new execution or a terminal status update.
	if active {
		s.upsertTaskThreadFromActionLocked(response, "")
	}
	if _, ok := s.agents[agent.AgentID]; ok {
		agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
		s.agents[agent.AgentID] = agent
	}
	s.removeTaskAgentMemberLocked(taskID, agent)
	if active {
		s.appendAgentTaskActionEvent(response)
	}
	return response, nil
}
