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
	receiptKey := taskRequestReceiptKey(principal, taskID, req)
	if receiptKey != "" {
		if response, ok := s.taskRequestReceipts[receiptKey]; ok {
			return response, nil
		}
	}
	if receiptKey != "" && req.intentGateRequired && req.AssignmentID == "" {
		req.AssignmentID = intentGateAssignmentIDPrefix + receiptKey
	}
	// Replaying a request after explicit unassignment must not re-add membership.
	if req.AssignmentID != "" {
		for _, thread := range s.taskThreads[taskID] {
			if thread.AssignmentID == req.AssignmentID && thread.AgentID == req.AgentID {
				return s.rememberTaskRequestReceiptLocked(receiptKey, actionResponseFromThread(thread, principal.WorkspaceID)), nil
			}
		}
	}
	s.assignTaskAgentMemberLocked(taskID, agent)
	if thread, ok := s.activeTaskThreadForAgentLocked(taskID, agent.AgentID); ok && canReuseActiveTaskThreadForAssignment(thread, req.AssignmentID) {
		return s.rememberTaskRequestReceiptLocked(receiptKey, actionResponseFromThread(thread, principal.WorkspaceID)), nil
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
	if req.conversationID != "" {
		s.adoptLegacyCommentConversationsLocked(taskID, req)
		sourceMessageID := req.SourceMessageID
		if sourceMessageID == "" {
			sourceMessageID = req.SourceCommentID
		}
		s.upsertTaskThreadMessageFromActionLocked(response, sourceMessageID, req.conversationID, req.parentThreadID)
		if !req.skipSourceMessage {
			s.appendThreadUserMessageLocked(response, principal, req.Instruction, req.SourceCommentID, req.sourceMessageObservedAt)
		}
	}
	s.upsertTaskThreadFromActionLocked(response, req.SourceCommentID)
	for i := range s.taskThreads[taskID] {
		if s.taskThreads[taskID][i].ThreadID == response.ThreadID {
			s.taskThreads[taskID][i].CommentInvocation = copyCommentInvocation(req.commentInvocation)
		}
	}
	agent = s.projectAgentWorkStatusFromThreadsLocked(agent)
	s.agents[agent.AgentID] = agent
	s.assignTaskAgentMemberLocked(taskID, agent)

	s.appendAgentTaskActionEvent(response)
	s.rememberTaskRequestReceiptLocked(receiptKey, response)
	return clientVisibleQueuedActionResponse(response), nil
}
