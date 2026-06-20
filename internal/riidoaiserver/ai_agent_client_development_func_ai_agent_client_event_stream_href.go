package riidoaiserver

import (
	"net/url"
	"strings"
)

func aiAgentClientEventStreamHref(workspaceID string) string {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return "/v1/client/ai-agent/events"
	}
	return "/v2/client/workspaces/" + url.PathEscape(workspaceID) + "/ai-agent/events"
}
