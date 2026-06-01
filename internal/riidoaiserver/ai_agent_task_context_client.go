package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	AIAgentTaskContextHeaderWorkspaceAPIKey = "X-Workspace-Api-Key"
	DefaultAIAgentTaskContextClientTimeout  = 2 * time.Second
)

type AIAgentTaskContextReader interface {
	GetAIAgentTaskContext(ctx context.Context, componentID string) (AIAgentTaskContext, error)
}

type AIAgentTaskContextClientConfig struct {
	BaseURL         string
	WorkspaceID     string
	TeamID          string
	WorkspaceAPIKey string
	Timeout         time.Duration
	HTTPClient      *http.Client
}

type AIAgentTaskContextClient struct {
	baseURL         string
	workspaceID     string
	teamID          string
	workspaceAPIKey string
	timeout         time.Duration
	httpClient      *http.Client
}

func NewAIAgentTaskContextClient(config AIAgentTaskContextClientConfig) (*AIAgentTaskContextClient, error) {
	baseURL, err := normalizeAIAgentTaskContextBaseURL(strings.TrimSpace(config.BaseURL))
	if err != nil {
		return nil, err
	}
	workspaceID := strings.TrimSpace(config.WorkspaceID)
	if workspaceID == "" {
		return nil, errors.New("riidoaiserver: task context workspace_id is required")
	}
	teamID := strings.TrimSpace(config.TeamID)
	if teamID == "" {
		return nil, errors.New("riidoaiserver: task context team_id is required")
	}
	workspaceAPIKey := strings.TrimSpace(config.WorkspaceAPIKey)
	if workspaceAPIKey == "" {
		return nil, errors.New("riidoaiserver: task context workspace api key is required")
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultAIAgentTaskContextClientTimeout
	}
	if timeout < 0 {
		return nil, errors.New("riidoaiserver: task context timeout must be positive")
	}
	return &AIAgentTaskContextClient{
		baseURL:         baseURL,
		workspaceID:     workspaceID,
		teamID:          teamID,
		workspaceAPIKey: workspaceAPIKey,
		timeout:         timeout,
		httpClient:      aiAgentTaskContextHTTPClient(config.HTTPClient, timeout),
	}, nil
}

func (c *AIAgentTaskContextClient) GetAIAgentTaskContext(ctx context.Context, componentID string) (AIAgentTaskContext, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskContext{}, err
	}
	if c == nil {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context client is not configured")
	}
	componentID = strings.TrimSpace(componentID)
	if componentID == "" {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context component_id is required")
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodGet, c.endpoint(componentID), nil)
	if err != nil {
		return AIAgentTaskContext{}, fmt.Errorf("riidoaiserver: create task context request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set(AIAgentTaskContextHeaderWorkspaceAPIKey, c.workspaceAPIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return AIAgentTaskContext{}, fmt.Errorf("riidoaiserver: task context request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return AIAgentTaskContext{}, fmt.Errorf("riidoaiserver: task context returned HTTP %d", resp.StatusCode)
	}

	var out AIAgentTaskContext
	dec := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return AIAgentTaskContext{}, fmt.Errorf("riidoaiserver: decode task context response: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: decode task context response: trailing data")
	}
	return out, nil
}

func (c *AIAgentTaskContextClient) endpoint(componentID string) string {
	return strings.TrimRight(c.baseURL, "/") +
		"/workspaces/" + url.PathEscape(c.workspaceID) +
		"/open-api/v1/teams/" + url.PathEscape(c.teamID) +
		"/components/" + url.PathEscape(componentID) +
		"/ai-agent-context"
}

func normalizeAIAgentTaskContextBaseURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", errors.New("riidoaiserver: task context base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse task context base url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("riidoaiserver: task context base url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("riidoaiserver: task context base url host is required")
	}
	if parsed.User != nil {
		return "", errors.New("riidoaiserver: task context base url must not include userinfo")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("riidoaiserver: task context base url must not include query or fragment")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String(), nil
}

func aiAgentTaskContextHTTPClient(client *http.Client, timeout time.Duration) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: timeout}
}
