package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) GetAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (result AIAgentTaskContext, err error) {
	defer func() {
		if err != nil && c != nil && c.locatorVersion == 2 {
			err = latestTaskDocumentError{taskReadScopeError{withFailureStage("read_scope", err)}}
		}
	}()
	if err := ctx.Err(); err != nil {
		return AIAgentTaskContext{}, err
	}
	if c == nil {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: private task context client is not configured")
	}
	componentID := strings.TrimSpace(req.ComponentID)
	if componentID == "" {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context component_id is required")
	}
	bearerToken := strings.TrimSpace(req.BearerToken)
	if bearerToken == "" {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context bearer token is required")
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if c.locatorVersion == 2 {
		req.ComponentID, req.BearerToken = componentID, bearerToken
		if err := req.DocumentSnapshot.validate(); err != nil {
			return AIAgentTaskContext{}, err
		}
		return c.getNativeTaskReadScope(requestCtx, req)
	}
	location, err := c.getComponentWorkspace(requestCtx, componentID, bearerToken)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	workspaceID := strings.TrimSpace(req.WorkspaceID)
	if workspaceID != "" && location.Team.Workspace.ID != workspaceID {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context workspace mismatch")
	}
	if location.Team.ID == "" {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context team_id is missing")
	}
	if err := req.DocumentSnapshot.validate(); err != nil {
		return AIAgentTaskContext{}, err
	}
	if req.metadataOnly {
		detail, err := c.getExplicitViewerMetadata(requestCtx, location.Team.ID, componentID, bearerToken)
		if err != nil {
			return AIAgentTaskContext{}, err
		}
		if detail.ID != componentID {
			return AIAgentTaskContext{}, ErrAuthorizationForbidden
		}
		return AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: detail.ID, Title: detail.Title, ComponentType: detail.ComponentType}, TeamID: location.Team.ID}, nil
	}
	if c.assignmentReadScope || req.capturedReadScope {
		req.WorkspaceID = location.Team.Workspace.ID
		req.ComponentID, req.BearerToken = componentID, bearerToken
		out, err := c.getTaskReadScope(requestCtx, req, location.Team.ID)
		if err != nil {
			return AIAgentTaskContext{}, latestTaskDocumentError{taskReadScopeError{withFailureStage("read_scope", err)}}
		}
		return out, nil
	}

	return c.getLegacyTaskDetailContext(requestCtx, req, location.Team.ID, componentID, bearerToken)
}
