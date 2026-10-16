package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

const onboardingCompletionRunIDPrefix = "run-onboarding-seed-"

func (s *DevelopmentAIAgentClientStore) completeAIAgentOnboardingTask(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest, fanout bool) (AIAgentTaskActionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.RequestID = strings.TrimSpace(req.RequestID)
	if taskID == "" {
		return AIAgentTaskActionResponse{}, errors.New("task_id is required")
	}
	if req.AgentID == "" {
		return AIAgentTaskActionResponse{}, errors.New("agent_id is required")
	}
	text, err := onboardingCompletionText(req.TemplateKey, req.Locale)
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, ok := s.visibleAgent(principal, req.AgentID)
	if !ok || agent.OwnerPrincipalID != principal.PrincipalID {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	receiptKey := onboardingCompletionReceiptKey(principal, req.RequestID)
	if receiptKey != "" {
		if response, ok := s.taskRequestReceipts[receiptKey]; ok {
			if response.TaskID != taskID || response.AgentID != agent.AgentID {
				return AIAgentTaskActionResponse{}, errOnboardingCompletionConflictForRequestID()
			}
			return response, nil
		}
	}
	assignmentID := onboardingSeedAssignmentID(taskID, agent.AgentID)
	if thread, ok := s.onboardingCompletionThreadLocked(taskID, agent.AgentID); ok {
		return s.rememberTaskRequestReceiptLocked(receiptKey, actionResponseFromThread(thread, principal.WorkspaceID)), nil
	}
	if err := s.checkOnboardingCompletionEligibleLocked(taskID, agent.AgentID); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	now := time.Now().UTC()
	s.assignTaskAgentMemberLocked(taskID, agent)
	runID := onboardingCompletionRunIDPrefix + taskID + "-" + agent.AgentID
	response := AIAgentTaskActionResponse{
		SchemaVersion:   SchemaVersion,
		TaskID:          taskID,
		AssignmentID:    assignmentID,
		AgentID:         agent.AgentID,
		AgentSnapshot:   s.agentSnapshotFromAgent(agent, now),
		RunID:           runID,
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
		Message:         text,
	}
	response.ThreadID = threadIDForRun(taskID, agent.AgentID, runID)
	s.upsertTaskThreadFromActionLocked(response, "")
	s.dropTaskThreadHistoryCacheLocked(response.ThreadID)
	s.appendAgentTaskActionEventWithFanout(response, fanout)
	stored, ok := s.onboardingSeedThreadLocked(taskID, assignmentID)
	if !ok {
		return AIAgentTaskActionResponse{}, errors.New("onboarding completion thread was not recorded")
	}
	s.markOnboardingCompletedLocked(agent.AgentID, taskID, stored)
	return s.rememberTaskRequestReceiptLocked(receiptKey, actionResponseFromThread(stored, principal.WorkspaceID)), nil
}
