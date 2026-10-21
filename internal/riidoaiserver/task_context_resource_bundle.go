package riidoaiserver

import (
	"context"
	"encoding/json"
)

const taskContextResourceQuery = `query AgentTaskResources($id: String!, $team: String!) {
 files: getComponentFiles(componentId:$id,teamId:$team) { id fileName fileLocation downloadUrl }
 links: getComponentWebLinks(componentId:$id,teamId:$team) { id linkName linkUrl }
}`

func (c *AIAgentPrivateTaskContextClient) readTaskContextResourceBundle(ctx context.Context, team, id, bearer string) ([]AIAgentTaskContextResource, string) {
	ctx, finish := observeTaskContextRead(ctx, "resource_metadata")
	outcome := "unavailable"
	defer func() { finish(outcome) }()
	var out struct {
		Files json.RawMessage `json:"files"`
		Links json.RawMessage `json:"links"`
	}
	if err := c.contextGraphQL(ctx, bearer, taskContextResourceQuery, team, id, &out); err != nil {
		outcome = taskContextReadOutcome(ctx, err)
		return nil, "unavailable"
	}
	files, filesOK := decodePrivateContextResources(out.Files, "file")
	links, linksOK := decodePrivateContextResources(out.Links, "link")
	outcome = "available"
	if !filesOK || !linksOK {
		outcome = "partial"
	}
	files = append(files, links...)
	if outcome == "partial" && len(files) == 0 {
		outcome = "unavailable"
	}
	return files, outcome
}
