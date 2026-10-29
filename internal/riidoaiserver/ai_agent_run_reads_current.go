package riidoaiserver

import (
	"context"
	"errors"
)

// Both HTTP history and SSE authorize against the current viewer response;
// retained content and service credentials must never grant visibility.
func (s Server) loadCurrentRunReadSources(ctx context.Context, token, workspaceID, taskID, assignmentID string) (map[runReadSourceKey]bool, error) {
	current, err := s.loadCurrentRunReadContext(ctx, token, workspaceID, taskID, assignmentID)
	if err != nil {
		return nil, err
	}
	return visibleRunReadSources(current), nil
}

func (s Server) loadCurrentRunReadContext(ctx context.Context, token, workspaceID, taskID, assignmentID string) (AIAgentTaskContext, error) {
	reader, ok := s.taskContext.(AIAgentTaskContextRequestReader)
	if !ok || token == "" {
		return AIAgentTaskContext{}, errors.New("source access unavailable")
	}
	scoped, err := s.currentRunReadRequiresScope(ctx, assignmentID, workspaceID, taskID)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	current, err := reader.GetAIAgentTaskContextForRequest(ctx, AIAgentTaskContextRequest{ComponentID: taskID, WorkspaceID: workspaceID, BearerToken: token, Extended: true, ReplyMetadata: true, capturedReadScope: scoped})
	if errors.Is(err, ErrAuthorizationForbidden) || taskContextReadOutcome(ctx, err) == "denied" {
		return AIAgentTaskContext{}, ErrAuthorizationForbidden
	}
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	if current.Component.ID != taskID {
		return AIAgentTaskContext{}, ErrAuthorizationForbidden
	}
	return current, ctx.Err()
}
