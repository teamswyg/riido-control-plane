package riidoaiserver

import "context"

func (c *AIAgentPrivateTaskContextClient) loadTaskReadScope(ctx context.Context, req AIAgentTaskContextRequest, team string) (taskReadScope, error) {
	query, variables, err := c.taskReadScopeRequest(req, team)
	if err != nil {
		return taskReadScope{}, err
	}
	var scope taskReadScope
	if err := c.readScopeGraphQL(ctx, req.BearerToken, query, "getComponentReadScope", variables, &scope); err != nil {
		return scope, err
	}
	if !scopeMatchesLocator(scope, c.locatorVersion, req.WorkspaceID) {
		return scope, taskContextReadError{"invalid"}
	}
	if c.locatorVersion == 2 && team == "" {
		// The authorized native root resolves its current team in this read.
		// Parent leaves may be central, but an assigned task must own a team.
		team = scope.TeamID
		if team == "" {
			return scope, taskContextReadError{"invalid"}
		}
	}
	return scope, validateTaskReadScope(scope, req.ComponentID, team)
}

func (c *AIAgentPrivateTaskContextClient) taskReadScopeRequest(req AIAgentTaskContextRequest, team string) (string, map[string]any, error) {
	variables := map[string]any{"componentId": req.ComponentID, "teamId": team, "editorSnapshot": nil}
	if req.DocumentSnapshot != nil {
		variables["editorSnapshot"] = map[string]string{"content": req.DocumentSnapshot.Content, "contentFormat": req.DocumentSnapshot.ContentFormat}
	}
	query := readScopeQuery
	if c.locatorVersion == 2 {
		if req.WorkspaceID == "" {
			return "", nil, taskContextReadError{"invalid"}
		}
		query = nativeReadScopeQuery(query)
		variables["workspaceId"], variables["locatorVersion"] = req.WorkspaceID, 2
		if team == "" {
			variables["teamId"] = nil
		}
	}
	return query, variables, nil
}
