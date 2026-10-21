package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func (c *AIAgentPrivateTaskContextClient) contextGraphQL(ctx context.Context, bearer, query, team, id string, out any) error {
	body, err := json.Marshal(map[string]any{"query": query, "variables": map[string]string{"team": team, "id": id}})
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
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := c.doPrivateTaskContextJSON(req, &envelope); err != nil {
		return err
	}
	for _, failure := range envelope.Errors {
		switch failure.Extensions.Code {
		case "UNAUTHENTICATED", "FORBIDDEN":
			return taskContextReadError{"denied"}
		default:
			return taskContextReadError{"unavailable"}
		}
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return taskContextReadError{"unavailable"}
	}
	if json.Unmarshal(envelope.Data, out) != nil {
		return taskContextReadError{"invalid"}
	}
	return nil
}
