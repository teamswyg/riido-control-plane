package riidoaiserver

import (
	"context"
	"errors"
)

// Snapshot rows and live messages must pass the same viewer-scoped task check.
func (s Server) dashboardTaskForViewer(ctx context.Context, token, workspaceID, taskID string) (AIAgentTaskContextComponent, error) {
	var empty AIAgentTaskContextComponent
	auth, err := s.config.Authorizer.Authorize(ctx, token, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, WorkspaceID: workspaceID, TaskID: taskID})
	if errors.Is(err, ErrAuthorizationForbidden) || (err == nil && auth.WorkspaceID != "" && auth.WorkspaceID != workspaceID) {
		return empty, ErrAuthorizationForbidden
	}
	if err != nil {
		return empty, errors.New("task access could not be verified")
	}
	reader, ok := s.taskContext.(AIAgentTaskContextRequestReader)
	if !ok {
		return empty, errors.New("request-scoped task context is unavailable")
	}
	read := reader.GetAIAgentTaskContextForRequest
	if metadata, ok := s.taskContext.(agentDashboardTaskReader); ok {
		read = metadata.GetAgentDashboardTask
	}
	task, err := read(ctx, AIAgentTaskContextRequest{ComponentID: taskID, WorkspaceID: workspaceID, BearerToken: token})
	if errors.Is(err, ErrAuthorizationForbidden) || errors.Is(err, errDashboardTaskUnavailable) {
		return empty, ErrAuthorizationForbidden
	}
	if err != nil {
		return empty, errors.New("task context could not be verified")
	}
	if task.Component.ID != taskID {
		return empty, ErrAuthorizationForbidden
	}
	return task.Component, nil
}
