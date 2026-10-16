package riidoaiserver

import (
	"context"
	"fmt"
	"strings"
)

type agentDashboardHistoryReader interface {
	LoadAgentDashboardHistory(context.Context, string) (AgentDashboardHistory, error)
}

// completeAIAgentOnboardingTaskUnderGate records the completion inside the task
// publication gate. An assignment holds the same gate from its durable write until its
// thread is published, so a completion sees either all of an assignment or none of it.
// It also refuses while the assignment store holds a non-terminal execution of this
// agent on this task.
func (s Server) completeAIAgentOnboardingTaskUnderGate(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest) (AIAgentTaskActionResponse, error) {
	release, err := s.taskAgentExecutionGate.beginPublication(ctx, taskID)
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	defer release()
	active, err := s.onboardingCompletionHasDurableExecution(ctx, taskID, strings.TrimSpace(req.AgentID))
	if err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	if active {
		return AIAgentTaskActionResponse{}, fmt.Errorf("%w: agent has a durable assignment on this task", errOnboardingCompletionConflict)
	}
	return s.aiAgent.CompleteAIAgentOnboardingTask(ctx, principal, taskID, req)
}

// onboardingCompletionHasDurableExecution reports whether the assignment store holds a
// non-terminal run of this agent on this task. A store without run history reports none.
func (s Server) onboardingCompletionHasDurableExecution(ctx context.Context, taskID, agentID string) (bool, error) {
	reader, ok := s.assignment.(agentDashboardHistoryReader)
	if !ok || agentID == "" {
		return false, nil
	}
	history, err := reader.LoadAgentDashboardHistory(ctx, agentID)
	if err != nil {
		return false, err
	}
	for _, run := range history.Runs {
		if run.Assignment.TaskID == taskID && run.Assignment.AgentID == agentID && !isTerminal(run.Assignment.State) {
			return true, nil
		}
	}
	return false, nil
}
