package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) readTaskContextResources(ctx context.Context, teamID, componentID, bearer string) ([]AIAgentTaskContextResource, string) {
	if c.projectionBundles {
		return c.readTaskContextResourceBundle(ctx, teamID, componentID, bearer)
	}
	var resources []AIAgentTaskContextResource
	status := "available"
	for _, item := range []struct{ path, kind string }{{"files", "file"}, {"web-links", "link"}} {
		endpoint := strings.Split(c.privateComponentDetailEndpoint(teamID, componentID), "?")[0] + "/" + item.path
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			status = "partial"
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+bearer)
		var data json.RawMessage
		if c.doPrivateTaskContextJSON(req, &data) != nil {
			status = "partial"
			continue
		}
		items, complete := decodePrivateContextResources(data, item.kind)
		if !complete {
			status = "partial"
		}
		resources = append(resources, items...)
	}
	if status == "partial" && len(resources) == 0 {
		status = "unavailable"
	}
	return resources, status
}
