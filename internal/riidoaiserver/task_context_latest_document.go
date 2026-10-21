package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// A latest-document failure must never become a successful run with stale HTML.
type latestTaskDocumentError struct{ cause error }

func (e latestTaskDocumentError) Error() string {
	return "latest task document unavailable; retry the same request_id"
}
func (e latestTaskDocumentError) Unwrap() error { return e.cause }

func (c *AIAgentPrivateTaskContextClient) getLatestProviderDocument(ctx context.Context, teamID, componentID, bearer string) (AIAgentTaskContextDocument, error) {
	endpoint := strings.TrimSuffix(c.privateProviderDocumentEndpoint(teamID, componentID), "html") + "json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return AIAgentTaskContextDocument{}, latestTaskDocumentError{err}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)
	var response struct {
		TiptapJSON json.RawMessage `json:"tiptapJson"`
	}
	if err := c.doPrivateTaskContextJSON(req, &response); err != nil {
		return AIAgentTaskContextDocument{}, latestTaskDocumentError{err}
	}
	if !providerTiptapJSONAbsent(response.TiptapJSON) {
		content, err := readableTaskContextTiptapJSON(response.TiptapJSON)
		if err != nil {
			return AIAgentTaskContextDocument{}, latestTaskDocumentError{err}
		}
		return AIAgentTaskContextDocument{Content: content, ContentFormat: "text", Source: "provider_document_json"}, nil
	}
	// Older documents may not have JSON yet. Only that explicit absence permits HTML.
	htmlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.privateProviderDocumentEndpoint(teamID, componentID), nil)
	if err != nil {
		return AIAgentTaskContextDocument{}, latestTaskDocumentError{err}
	}
	htmlReq.Header = req.Header.Clone()
	var responseHTML privateProviderDocumentResponse
	if err := c.doPrivateTaskContextJSON(htmlReq, &responseHTML); err != nil {
		return AIAgentTaskContextDocument{}, latestTaskDocumentError{err}
	}
	return AIAgentTaskContextDocument{Content: responseHTML.HTML, ContentFormat: "html", Source: "provider_document"}, nil
}

func providerTiptapJSONAbsent(raw json.RawMessage) bool {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return true
	}
	var serialized string
	return json.Unmarshal(raw, &serialized) == nil && serialized == ""
}
