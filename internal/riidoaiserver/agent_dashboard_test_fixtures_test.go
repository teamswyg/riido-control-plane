package riidoaiserver

import (
	"context"
	"errors"
)

type dashboardTestAuthorizer struct{ fail bool }

func (a dashboardTestAuthorizer) Authorize(_ context.Context, _ string, req AuthorizationRequest) (AuthorizationResult, error) {
	if a.fail && req.TaskID != "" {
		return AuthorizationResult{}, errors.New("temporary authorization failure")
	}
	if req.TaskID == "secret" {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	return AuthorizationResult{PrincipalID: "user-1", WorkspaceID: req.WorkspaceID}, nil
}

type dashboardTestContext struct{ calls []string }

func (c *dashboardTestContext) GetAIAgentTaskContext(context.Context, string) (AIAgentTaskContext, error) {
	panic("must not use service credentials")
}

func (c *dashboardTestContext) GetAIAgentTaskContextForRequest(_ context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	c.calls = append(c.calls, req.ComponentID)
	if req.BearerToken != "viewer-token" {
		panic("missing viewer credentials")
	}
	return AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: req.ComponentID, Title: "Visible task", KeyNumber: "RIID-5830"}}, nil
}
