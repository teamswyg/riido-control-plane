package riidoaiserver

import (
	"context"
	"fmt"
	"strings"
)

type onboardingCompletionRecordChecker interface {
	hasOnboardingCompletionRecord(taskID, agentID string) bool
}

// completeAIAgentOnboardingTaskUnderGate records the completion inside the task
// publication gate. An assignment holds the same gate from its durable write until its
// thread is published, so a completion sees either all of an assignment or none of it.
// A repeat of an existing completion is answered by the store with the recorded
// response. A new completion is refused while the assignment store has any run of this
// agent on this task, including a finished one.
func (s Server) completeAIAgentOnboardingTaskUnderGate(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest) (AIAgentTaskActionResponse, error) {
	release, err := s.taskAgentExecutionGate.beginPublication(ctx, taskID)
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	defer release()
	agentID := strings.TrimSpace(req.AgentID)
	if !s.onboardingCompletionRecorded(taskID, agentID) {
		hasRun, err := s.onboardingCompletionHasDurableRun(ctx, taskID, agentID)
		if err != nil {
			return AIAgentTaskActionResponse{}, err
		}
		if hasRun {
			return AIAgentTaskActionResponse{}, fmt.Errorf("%w: agent has a durable assignment on this task", errOnboardingCompletionConflict)
		}
	}
	return s.aiAgent.CompleteAIAgentOnboardingTask(ctx, principal, taskID, req)
}

// onboardingCompletionRecorded reports whether this agent already has a completion
// record on this task. The store answers such a repeat itself, so later runs must not
// block it.
func (s Server) onboardingCompletionRecorded(taskID, agentID string) bool {
	checker, ok := s.aiAgent.(onboardingCompletionRecordChecker)
	return ok && checker.hasOnboardingCompletionRecord(taskID, agentID)
}

func (s *DevelopmentAIAgentClientStore) hasOnboardingCompletionRecord(taskID, agentID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.onboardingCompletions[agentID] == taskID
}

// onboardingCompletionHasDurableRun reports whether the assignment store holds any run
// of this agent on this task. A store without run history reports none.
func (s Server) onboardingCompletionHasDurableRun(ctx context.Context, taskID, agentID string) (bool, error) {
	reader, ok := s.assignment.(AgentDashboardHistoryReader)
	if !ok || agentID == "" {
		return false, nil
	}
	history, err := reader.LoadAgentDashboardHistory(ctx, agentID)
	if err != nil {
		return false, assignmentDependencyFailure(err)
	}
	for _, run := range history.Runs {
		if run.Assignment.TaskID == taskID && run.Assignment.AgentID == agentID {
			return true, nil
		}
	}
	return false, nil
}
