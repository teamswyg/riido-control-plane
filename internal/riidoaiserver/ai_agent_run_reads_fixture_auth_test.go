package riidoaiserver

import (
	"context"
)

type runReadsTestContext struct {
	value    AIAgentTaskContext
	err      error
	requests []AIAgentTaskContextRequest
}

func (*runReadsTestContext) GetAIAgentTaskContext(context.Context, string) (AIAgentTaskContext, error) {
	panic("viewer reads must not use service credentials")
}

func (c *runReadsTestContext) GetAIAgentTaskContextForRequest(_ context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	c.requests = append(c.requests, req)
	return c.value, c.err
}

type runReadsTestAuthorizer struct {
	fail   bool
	tokens []string
}

func (a *runReadsTestAuthorizer) Authorize(_ context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
	a.tokens = append(a.tokens, token)
	if a.fail && req.TaskID != "" {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	workspace := req.WorkspaceID
	if workspace == "" {
		workspace = defaultAIAgentClientWorkspaceID
	}
	return AuthorizationResult{PrincipalID: "viewer", WorkspaceID: workspace}, nil
}
