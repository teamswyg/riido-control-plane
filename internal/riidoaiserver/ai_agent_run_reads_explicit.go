package riidoaiserver

import (
	"context"
	"errors"
)

func explicitReadSource(key runReadSourceKey) bool {
	if key.kind != "task" && key.kind != "project" && key.kind != "milestone" {
		return false
	}
	return key.part == "metadata" && key.id != ""
}

// A receipt is not an access grant. Resolve its owner using this viewer's
// Bearer on every HTTP/SSE delivery; never use the run's service credential.
func (s Server) currentExplicitReadMetadata(ctx context.Context, token, workspace string, key runReadSourceKey) (runReadMetadata, bool, error) {
	empty := runReadMetadata{}
	if !explicitReadSource(key) {
		return empty, false, nil
	}
	reader, ok := s.taskContext.(AIAgentTaskContextRequestReader)
	if !ok || token == "" {
		return empty, false, errors.New("source access unavailable")
	}
	current, err := reader.GetAIAgentTaskContextForRequest(ctx, AIAgentTaskContextRequest{ComponentID: key.id, WorkspaceID: workspace, BearerToken: token, metadataOnly: true})
	if errors.Is(err, ErrAuthorizationForbidden) || (taskContextReadOutcome(ctx, err) == "denied" || taskContextReadOutcome(ctx, err) == "missing") {
		return empty, false, nil
	}
	if err != nil {
		return empty, false, err
	}
	kind := current.Component.ComponentType
	if kind == "subtask" {
		kind = "task"
	}
	if current.Component.ID != key.id || kind != key.kind {
		return empty, false, nil
	}
	return runReadMetadata{title: current.Component.Title, source: AssignmentContextSource{Kind: key.kind, ID: key.id, Part: key.part}}, true, ctx.Err()
}

func (c *AIAgentPrivateTaskContextClient) getExplicitViewerMetadata(ctx context.Context, team, id, bearer string) (AIAgentTaskContextComponent, error) {
	var out struct {
		Metadata AIAgentTaskContextComponent `json:"metadata"`
	}
	const query = `query AgentExplicitReadMetadata($id:String!,$team:String!){metadata:getComponent(componentId:$id,teamId:$team,getComponentDto:{getDocument:"false"}){id title componentType}}`
	err := c.contextGraphQL(ctx, bearer, query, team, id, &out)
	return out.Metadata, err
}
