package riidoaiserver

import (
	"context"
	"errors"
	"strconv"
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
	if thread, ok := s.activeTaskThreadForAgentLocked(taskID, agent.AgentID); ok {
		return actionResponseFromThread(thread, principal.WorkspaceID), nil
	}
	if stopExistingTaskThreads {
		s.markTaskActiveThreadsStoppedLocked(taskID, AgentTaskCommentStoppedByUserRequest, "agent assignment was replaced by a participant change")
	}
	sequence := strconv.Itoa(len(s.taskThreads[taskID]) + 1)
	if req.AssignmentID == "" {
		req.AssignmentID = "asn-dev-assignment-" + taskID + "-" + sequence
	}
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
		Message:         "agent assignment started from task participant",
	}
	response.ThreadID = threadIDForRun(response.TaskID, response.AgentID, response.RunID)
	if agent.WorkStatus == AgentWorkStatusRunning || agent.WorkStatus == AgentWorkStatusWaitingForUser || agent.WorkStatus == AgentWorkStatusQueued {
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		response.Message = "agent is busy; task assignment was queued"
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
