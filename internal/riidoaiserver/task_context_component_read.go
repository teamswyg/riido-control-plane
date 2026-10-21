package riidoaiserver

import (
	"context"
	"net/http"
)

func (c *AIAgentPrivateTaskContextClient) getComponentWorkspace(ctx context.Context, id, bearer string) (privateComponentWorkspaceResponse, error) {
	var out privateComponentWorkspaceResponse
	if err := c.readPrivateContextComponent(ctx, c.privateComponentWorkspaceEndpoint(id), bearer, "workspace_location", &out); err != nil {
		return out, err
	}
	return out.normalized(), nil
}

func (c *AIAgentPrivateTaskContextClient) getComponentDetail(ctx context.Context, team, id, bearer string) (privateComponentDetailResponse, error) {
	var out privateComponentDetailResponse
	if err := c.readPrivateContextComponent(ctx, c.privateComponentDetailEndpoint(team, id), bearer, "component_detail", &out); err != nil {
		return out, err
	}
	return out.normalized(), nil
}

func (c *AIAgentPrivateTaskContextClient) readPrivateContextComponent(ctx context.Context, endpoint, bearer, operation string, out any) (err error) {
	ctx, finish := observeTaskContextRead(ctx, operation)
	defer func() { finish(taskContextReadOutcome(ctx, err)) }()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return taskContextReadError{"invalid"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)
	return c.doPrivateTaskContextJSON(req, out)
}
