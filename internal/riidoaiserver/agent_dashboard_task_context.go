package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type agentDashboardTaskReader interface {
	GetAgentDashboardTask(context.Context, AIAgentTaskContextRequest) (AIAgentTaskContext, error)
}

// Progress needs only identity and display metadata. Use the current monolith
// GraphQL contract with the viewer token, without fetching task documents.
func (c *AIAgentPrivateTaskContextClient) GetAgentDashboardTask(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	var location struct {
		Component privateComponentWorkspaceResponse `json:"getComponentWorkspace"`
	}
	err := c.dashboardGraphQL(ctx, req.BearerToken, `query AgentDashboardLocation($id: String!) { getComponentWorkspace(componentId: $id) { id componentType team { id workspace { id } } } }`, map[string]string{"id": req.ComponentID}, &location)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	component := location.Component
	if component.ID != req.ComponentID || component.Team.Workspace.ID != req.WorkspaceID || component.Team.ID == "" || (component.ComponentType != "task" && component.ComponentType != "subtask") {
		return AIAgentTaskContext{}, ErrAuthorizationForbidden
	}
	var detail struct {
		Component AIAgentTaskContextComponent `json:"getComponent"`
	}
	err = c.dashboardGraphQL(ctx, req.BearerToken, `query AgentDashboardTask($id: String!, $team: String!) { getComponent(componentId: $id, teamId: $team, getComponentDto: {getDocument: "false"}) { id componentType title keyNumber } }`, map[string]string{"id": req.ComponentID, "team": component.Team.ID}, &detail)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	if detail.Component.ID != req.ComponentID || (detail.Component.ComponentType != "task" && detail.Component.ComponentType != "subtask") {
		return AIAgentTaskContext{}, ErrAuthorizationForbidden
	}
	return AIAgentTaskContext{Component: detail.Component, TeamID: component.Team.ID}, nil
}

func (c *AIAgentPrivateTaskContextClient) dashboardGraphQL(ctx context.Context, token, query string, variables map[string]string, out any) error {
	if strings.TrimSpace(token) == "" {
		return ErrAuthorizationForbidden
	}
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}
	endpoint := strings.TrimRight(c.graphqlURL, "/")
	if !strings.HasSuffix(endpoint, "/graphql") {
		endpoint += "/graphql"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dashboard metadata request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusForbidden {
		return ErrAuthorizationForbidden
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("dashboard metadata returned HTTP %d", res.StatusCode)
	}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&envelope); err != nil {
		return err
	}
	for _, failure := range envelope.Errors {
		switch failure.Extensions.Code {
		case "FORBIDDEN", "NOT_FOUND":
			return ErrAuthorizationForbidden
		default:
			return fmt.Errorf("dashboard metadata query failed")
		}
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return fmt.Errorf("dashboard metadata response missing data")
	}
	return json.Unmarshal(envelope.Data, out)
}
