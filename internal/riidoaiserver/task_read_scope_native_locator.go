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
	// TASK reads require the current team. If a move races these two reads,
	// the producer reauthorizes and the scope/root checks reject a stale team.
	return c.getTaskReadScope(ctx, req, location.Team.ID)
}
