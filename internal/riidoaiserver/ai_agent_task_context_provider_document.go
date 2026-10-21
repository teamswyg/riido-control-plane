package riidoaiserver

import (
	"net/url"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) privateProviderDocumentEndpoint(teamID, componentID string) string {
	return strings.TrimRight(c.baseURL, "/") +
		"/documents/providers/" + url.PathEscape(teamID) +
		"/" + url.PathEscape(componentID) +
		"?format=html"
}
