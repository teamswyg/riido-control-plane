package riidoaiserver

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) getProviderDocumentHTML(ctx context.Context, teamID, componentID, bearerToken string) (privateProviderDocumentResponse, bool) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.privateProviderDocumentEndpoint(teamID, componentID), nil)
	if err != nil {
		return privateProviderDocumentResponse{}, false
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+bearerToken)
	var out privateProviderDocumentResponse
	if err := c.doPrivateTaskContextJSON(httpReq, &out); err != nil {
		return privateProviderDocumentResponse{}, false
	}
	out = out.normalized()
	if out.HTML == "" {
		return privateProviderDocumentResponse{}, false
	}
	return out, true
}

func (c *AIAgentPrivateTaskContextClient) privateProviderDocumentEndpoint(teamID, componentID string) string {
	return strings.TrimRight(c.baseURL, "/") +
		"/documents/providers/" + url.PathEscape(teamID) +
		"/" + url.PathEscape(componentID) +
		"?format=html"
}

type privateProviderDocumentResponse struct {
	HTML string `json:"html"`
}

func (r privateProviderDocumentResponse) normalized() privateProviderDocumentResponse {
	r.HTML = strings.TrimSpace(r.HTML)
	return r
}
