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
	if s.stopTargetsOnboardingCompletionLocked(taskID, agent.AgentID, req.AssignmentID) {
		return AIAgentTaskActionResponse{}, errOnboardingCompletionStopConflict
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
		return AIAgentTaskActionResponse{}, errors.New("active ai agent assignment not found")
	}
	if req.AssignmentID != "" {
		s.markTaskAgentAssignmentThreadStopProjectionLocked(taskID, agent.AgentID, req.AssignmentID, response, completed)
	} else {
		s.markTaskAgentThreadsStopProjectionLocked(taskID, agent.AgentID, response, completed)
	}
	response = actionResponseWithActiveStream(response, principal.WorkspaceID)
	s.upsertTaskThreadFromActionLocked(response, "")
	if _, ok := s.agents[agent.AgentID]; ok {
		agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
		s.agents[agent.AgentID] = agent
	}
	s.appendAgentTaskActionEvent(response)
	return response, nil
}

// stopTargetsOnboardingCompletionLocked reports whether a stop would land on the
// onboarding completion record: the named assignment is onboarding-seed, or the
// resolved target thread is one, or no target is named and the only record for
// this agent on the task is the onboarding completion.
func (s *DevelopmentAIAgentClientStore) stopTargetsOnboardingCompletionLocked(taskID, agentID, assignmentID string) bool {
	if isOnboardingSeedAssignmentID(assignmentID) {
		return true
	}
	if thread, ok := s.taskThreadForStopTargetLocked(taskID, agentID, assignmentID); ok {
		return isOnboardingSeedAssignmentID(thread.AssignmentID)
	}
	if assignmentID != "" {
		return false
	}
	_, ok := s.onboardingSeedThreadLocked(taskID, onboardingSeedAssignmentID(taskID, agentID))
	return ok
}
