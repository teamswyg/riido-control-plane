package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
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
	targetAgentID := thread.AgentID
	if req.targetAgentID != "" {
		targetAgentID = req.targetAgentID
	}
	agent, ok := s.visibleAgent(principal, targetAgentID)
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	receiptKey := taskThreadMessageReceiptKey(principal, taskID, threadID, targetAgentID, req)
	if response, found := s.taskRequestReceipts[receiptKey]; receiptKey != "" && found {
		return response, nil
	}
	if req.SourceMessageID == "" && receiptKey != "" {
		req.SourceMessageID = "thread-message:" + receiptKey
	}

	response := taskThreadMessageInitialResponse(taskID, threadID, thread, agent)
	handoff := targetAgentID != thread.AgentID
	if handoff {
		response.AgentSnapshot = s.agentSnapshotFromAgent(agent, time.Now().UTC())
	}
	threadWasActive := !handoff && taskThreadHasActiveStream(thread)
	conversationID := taskThreadConversationID(thread)
	parentThreadID := ""
	if !req.toolApproval && (handoff || taskThreadMessageStartsNewExecution(thread)) {
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
	hideQueuedThreadReplyMessage(&response)
	response = actionResponseWithActiveStream(response, principal.WorkspaceID)
	s.appendThreadUserMessageLocked(response, principal, req.Body, req.SourceMessageID)
	s.upsertTaskThreadMessageFromActionLocked(response, req.SourceMessageID, conversationID, parentThreadID)
	if !req.toolApproval && !req.toolApprovalWithoutPending {
		s.assignTaskAgentMemberLocked(taskID, agent)
	}
	s.appendAgentTaskActionEvent(response)
	s.appendTaskNotificationLocked(TaskEvent{EventID: "started:" + response.AssignmentID, AssignmentID: response.AssignmentID, TaskID: taskID, AgentID: response.AgentID, WorkspaceID: principal.WorkspaceID, CreatedBy: principal.PrincipalID, At: time.Now().UTC()}, response)
	return s.rememberTaskRequestReceiptLocked(receiptKey, clientVisibleQueuedActionResponse(response)), nil
}
