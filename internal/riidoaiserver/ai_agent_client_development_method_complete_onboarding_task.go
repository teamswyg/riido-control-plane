package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"time"
)

const onboardingCompletionRunIDPrefix = "run-onboarding-seed-"

// CompleteAIAgentOnboardingTask adds the onboarding agent as a task participant and
// records one completed, agent-authored task_completed comment. It never calls the
// assignment queue and never starts execution.
func (s *DevelopmentAIAgentClientStore) CompleteAIAgentOnboardingTask(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest) (AIAgentTaskActionResponse, error) {
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
	if !ok {
		return AIAgentTaskActionResponse{}, ErrAIAgentNotFound
	}
	receiptKey := taskRequestReceiptKey(principal, taskID, AssignAIAgentTaskRequest{AgentID: req.AgentID, RequestID: req.RequestID})
	if receiptKey != "" {
		if response, ok := s.taskRequestReceipts[receiptKey]; ok {
			return response, nil
		}
	}
	assignmentID := onboardingSeedAssignmentID(taskID, agent.AgentID)
	if thread, ok := s.onboardingSeedThreadLocked(taskID, assignmentID); ok {
		return s.rememberTaskRequestReceiptLocked(receiptKey, actionResponseFromThread(thread, principal.WorkspaceID)), nil
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
	s.appendAgentTaskActionEvent(response)
	stored, ok := s.onboardingSeedThreadLocked(taskID, assignmentID)
	if !ok {
		return AIAgentTaskActionResponse{}, errors.New("onboarding completion thread was not recorded")
	}
	return s.rememberTaskRequestReceiptLocked(receiptKey, actionResponseFromThread(stored, principal.WorkspaceID)), nil
}

func (s *DevelopmentAIAgentClientStore) onboardingSeedThreadLocked(taskID, assignmentID string) (AIAgentTaskThreadRecord, bool) {
	for _, thread := range s.taskThreads[taskID] {
		if thread.AssignmentID == assignmentID {
			return thread, true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
