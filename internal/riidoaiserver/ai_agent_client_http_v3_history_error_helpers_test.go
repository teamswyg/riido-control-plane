package riidoaiserver

import (
	"context"
)

type v3HistoryErrorStore struct {
	*DevelopmentAIAgentClientStore
	historyErr   error
	reconcileErr error
}

func (s v3HistoryErrorStore) ListAIAgentTaskThreadHistory(
	context.Context,
	AuthorizationResult,
	string,
) (AIAgentTaskThreadHistoryCollectionResponse, error) {
	return AIAgentTaskThreadHistoryCollectionResponse{}, s.historyErr
}

func (s v3HistoryErrorStore) ReconcileAIAgentActiveThreadProjections(
	context.Context,
	AuthorizationResult,
	string,
	AssignmentProjectionReader,
) (bool, error) {
	return false, s.reconcileErr
}

func v3HistoryErrorTestPath() string {
	return "/v3/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-history-read/threads"
}
