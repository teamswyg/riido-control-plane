package riidoaiserver

import (
	"context"
	"errors"
)

// Both HTTP history and SSE authorize against the current viewer response;
// retained content and service credentials must never grant visibility.
func (s Server) loadCurrentRunReadSources(ctx context.Context, token, workspaceID, taskID string) (map[runReadSourceKey]bool, error) {
	reader, ok := s.taskContext.(AIAgentTaskContextRequestReader)
	if !ok || token == "" {
		return nil, errors.New("source access unavailable")
	}
	current, err := reader.GetAIAgentTaskContextForRequest(ctx, AIAgentTaskContextRequest{ComponentID: taskID, WorkspaceID: workspaceID, BearerToken: token, Extended: true})
	if errors.Is(err, ErrAuthorizationForbidden) || taskContextReadOutcome(ctx, err) == "denied" {
		return nil, ErrAuthorizationForbidden
	}
	if err != nil {
		return nil, err
	}
	if current.Component.ID != taskID {
		return nil, ErrAuthorizationForbidden
	}
	return visibleRunReadSources(current), ctx.Err()
}
