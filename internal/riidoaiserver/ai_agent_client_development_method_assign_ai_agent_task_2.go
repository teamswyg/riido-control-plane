package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) assignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest, stopExistingTaskThreads bool) (AIAgentTaskActionResponse, error) {
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
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	if thread, ok := s.activeTaskThreadForAgentLocked(taskID, agent.AgentID); ok && canReuseActiveTaskThreadForAssignment(thread, req.AssignmentID) {
		return actionResponseFromThread(thread, principal.WorkspaceID), nil
	}
	if stopExistingTaskThreads {
		s.markTaskActiveThreadsStoppedLocked(taskID, AgentTaskCommentStoppedByUserRequest, clientMessageTaskStopped)
		if refreshed, ok := s.visibleAgent(principal, agent.AgentID); ok {
			agent = refreshed
		}
	}
	sequence := nextTaskThreadSequence(s.taskThreads[taskID])
	req.AssignmentID = taskAssignmentIDForRequest(taskID, sequence, req)
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AssignmentID:    req.AssignmentID,
		AgentID:         agent.AgentID,
		AgentSnapshot:   s.agentSnapshotFromAgent(agent, time.Now().UTC()),
		RunID:           "run-dev-assignment-" + taskID + "-" + sequence,
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		CommentKind:     AgentTaskCommentAssignmentStarted,
		Message:         clientMessageTaskRunning,
	}
	response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	applyAssignmentStartProjection(&response, agent.WorkStatus, req)
	if assignmentStateIsKnown(req.durableState) {
		response.Message = ""
		applyAssignmentStateActionResponse(&response, req.durableState)
	}
	response = actionResponseWithActiveStream(response, principal.WorkspaceID)
	agent.WorkStatus = response.WorkStatus
	agent.Editability = AgentEditabilityBlockedAssignedTasks
	agent.AssignedTaskCount++
	s.agents[agent.AgentID] = agent
	s.upsertTaskThreadFromActionLocked(response, "")
	s.appendAgentTaskActionEvent(response)
	return response, nil
}
