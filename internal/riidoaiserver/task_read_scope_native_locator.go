package riidoaiserver

import "context"

const nativeTaskReadScopeLocationQuery = `query RiidoNativeTaskLocation($id:String!){getComponentWorkspace(componentId:$id){id componentType team{id workspace{id}}}}`

func (c *AIAgentPrivateTaskContextClient) getNativeTaskReadScope(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	if req.WorkspaceID == "" {
		return AIAgentTaskContext{}, taskContextReadError{"invalid"}
	}
	var location privateComponentWorkspaceResponse
	err := c.readScopeGraphQL(ctx, req.BearerToken, nativeTaskReadScopeLocationQuery, "getComponentWorkspace", map[string]string{"id": req.ComponentID}, &location)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	location = location.normalized()
	if location.ID != req.ComponentID || location.Team.Workspace.ID != req.WorkspaceID || location.Team.ID == "" || (location.ComponentType != "task" && location.ComponentType != "subtask") {
		return AIAgentTaskContext{}, taskContextReadError{"invalid"}
	}
	if req.metadataOnly {
		var detail AIAgentTaskContextComponent
		const query = `query AgentExplicitNativeMetadata($id:String!,$team:String!,$workspace:String!){metadata:getComponent(componentId:$id,teamId:$team,workspaceId:$workspace,locatorVersion:2,getComponentDto:{getDocument:"false"}){id title componentType}}`
		err := c.readScopeGraphQL(ctx, req.BearerToken, query, "metadata", map[string]string{"id": req.ComponentID, "team": location.Team.ID, "workspace": req.WorkspaceID}, &detail)
		if err != nil {
			return AIAgentTaskContext{}, err
		}
		if detail.ID != req.ComponentID {
			return AIAgentTaskContext{}, ErrAuthorizationForbidden
		}
		return AIAgentTaskContext{Component: detail, TeamID: location.Team.ID}, nil
	}
	// TASK reads require the current team. If a move races these two reads,
	// the producer reauthorizes and the scope/root checks reject a stale team.
	return c.getTaskReadScope(ctx, req, location.Team.ID)
}
