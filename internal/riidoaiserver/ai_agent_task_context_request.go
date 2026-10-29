package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) GetAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
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
	if c.assignmentReadScope || req.capturedReadScope {
		req.ComponentID, req.BearerToken = componentID, bearerToken
		out, err := c.getTaskReadScope(requestCtx, req, location.Team.ID)
		if err != nil {
			return AIAgentTaskContext{}, latestTaskDocumentError{err}
		}
		return out, nil
	}

	detail, err := c.getComponentDetail(requestCtx, location.Team.ID, componentID, bearerToken)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	if detail.ID != componentID {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context component identity mismatch")
	}
	contextSnapshot := detail.toAIAgentTaskContext()
	contextSnapshot.TeamID = location.Team.ID
	if err := req.DocumentSnapshot.validate(); err != nil {
		return AIAgentTaskContext{}, err
	}
	if err := c.enrichTaskContextFor(requestCtx, location.Team.ID, componentID, bearerToken, &contextSnapshot, req); err != nil {
		return AIAgentTaskContext{}, err
	}
	return contextSnapshot, nil
}
