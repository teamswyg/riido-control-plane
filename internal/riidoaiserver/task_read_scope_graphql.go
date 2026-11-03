package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// Queries are package-owned constants; callers supply variables, never queries.
func (c *AIAgentPrivateTaskContextClient) readScopeGraphQL(ctx context.Context, bearer, query, root string, variables, out any) error {
	if c.readScopeEvidence != nil {
		return c.readScopeEvidence.read(ctx, root, variables, out)
	}
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return taskContextReadError{"invalid"}
	}
	endpoint := strings.TrimRight(c.graphqlURL, "/")
	if !strings.HasSuffix(endpoint, "/graphql") {
		endpoint += "/graphql"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return taskContextReadError{"invalid"}
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Content-Type", "application/json")
	var envelope struct {
		Data   map[string]json.RawMessage `json:"data"`
		Errors []struct {
			Extensions struct {
				Code   string `json:"code"`
				Reason string `json:"reason"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := c.doPrivateTaskContextJSON(req, &envelope); err != nil {
		return err
	}
	if len(envelope.Errors) > 0 {
		switch envelope.Errors[0].Extensions.Code {
		case "UNAUTHENTICATED", "FORBIDDEN", "authentication_required", "permission_denied":
			return taskContextReadError{"denied"}
		case "not_found":
			return taskContextReadError{"missing"}
		case "conflict":
			// RIID-6008: stored data the monolith cannot serve safely. Retrying
			// the same request will not fix it.
			if envelope.Errors[0].Extensions.Reason == readScopeIntegrityReason {
				return taskContextReadError{"integrity"}
			}
		}
		return taskContextReadError{"unavailable"}
	}
	data := envelope.Data[root]
	if len(data) == 0 || string(data) == "null" || json.Unmarshal(data, out) != nil {
		return taskContextReadError{"invalid"}
	}
	return nil
}
