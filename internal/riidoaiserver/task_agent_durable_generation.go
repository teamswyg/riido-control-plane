package riidoaiserver

import (
	"context"
)

type taskAgentGenerationScope struct {
	taskID, agentID, generation string
}

type taskAgentMembershipSnapshotMutation struct {
	scope    taskAgentGenerationScope
	baseline AIAgentTaskMembership
}

type taskAgentSnapshotContextKey struct{}

type taskAgentPreserveMembershipContextKey struct{}

type taskAgentFreshSnapshotContextKey struct{}

func withTaskAgentSnapshotMutation(ctx context.Context, mutation taskAgentMembershipSnapshotMutation) context.Context {
	return context.WithValue(ctx, taskAgentSnapshotContextKey{}, mutation)
}

func taskAgentSnapshotMutation(ctx context.Context) (taskAgentMembershipSnapshotMutation, bool) {
	value, ok := ctx.Value(taskAgentSnapshotContextKey{}).(taskAgentMembershipSnapshotMutation)
	return value, ok
}

type taskAgentDurableExecutionReader interface {
	loadTaskAgentDurableExecutions(context.Context, string, string) (string, []AssignmentProjection, error)
}

func (s *Store) taskAgentGenerationSupported() bool {
	_, ok := s.operationStore.(taskAgentDurableExecutionReader)
	return ok
}

func (s *Store) loadTaskAgentDurableExecutions(ctx context.Context, taskID, agentID string) (string, []AssignmentProjection, error) {
	reader, ok := s.operationStore.(taskAgentDurableExecutionReader)
	if !ok {
		return "", nil, nil
	}
	return reader.loadTaskAgentDurableExecutions(ctx, taskID, agentID)
}
