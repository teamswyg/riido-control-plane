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

type AIAgentTaskContextRequest struct {
	ComponentID string
	WorkspaceID string
	BearerToken string
}

type AIAgentTaskContextRequestReader interface {
	GetAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error)
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

type AIAgentPrivateTaskContextClientConfig struct {
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type AIAgentPrivateTaskContextClient struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

func NewAIAgentPrivateTaskContextClient(config AIAgentPrivateTaskContextClientConfig) (*AIAgentPrivateTaskContextClient, error) {
	baseURL, err := normalizeAIAgentTaskContextBaseURL(strings.TrimSpace(config.BaseURL))
	if err != nil {
		return nil, err
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultAIAgentTaskContextClientTimeout
	}
	if timeout < 0 {
		return nil, errors.New("riidoaiserver: private task context timeout must be positive")
	}
	return &AIAgentPrivateTaskContextClient{
		baseURL:    baseURL,
		timeout:    timeout,
		httpClient: aiAgentTaskContextHTTPClient(config.HTTPClient, timeout),
	}, nil
}

func (c *AIAgentPrivateTaskContextClient) GetAIAgentTaskContext(ctx context.Context, componentID string) (AIAgentTaskContext, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskContext{}, err
	}
	return AIAgentTaskContext{}, errors.New("riidoaiserver: private task context requires request-scoped bearer token")
}

func (c *AIAgentPrivateTaskContextClient) GetAIAgentTaskContextForRequest(ctx context.Context, req AIAgentTaskContextRequest) (AIAgentTaskContext, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskContext{}, err
	}
	if c == nil {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: private task context client is not configured")
	}
	componentID := strings.TrimSpace(req.ComponentID)
	if componentID == "" {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context component_id is required")
	}
	bearerToken := strings.TrimSpace(req.BearerToken)
	if bearerToken == "" {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context bearer token is required")
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	location, err := c.getComponentWorkspace(requestCtx, componentID, bearerToken)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	workspaceID := strings.TrimSpace(req.WorkspaceID)
	if workspaceID != "" && location.Team.Workspace.ID != workspaceID {
		return AIAgentTaskContext{}, fmt.Errorf("riidoaiserver: task context workspace mismatch %q", location.Team.Workspace.ID)
	}
	if location.Team.ID == "" {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context team_id is missing")
	}

	detail, err := c.getComponentDetail(requestCtx, location.Team.ID, componentID, bearerToken)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	return detail.toAIAgentTaskContext(), nil
}

func (c *AIAgentPrivateTaskContextClient) getComponentWorkspace(ctx context.Context, componentID, bearerToken string) (privateComponentWorkspaceResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.privateComponentWorkspaceEndpoint(componentID), nil)
	if err != nil {
		return privateComponentWorkspaceResponse{}, fmt.Errorf("riidoaiserver: create task context workspace request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+bearerToken)
	var out privateComponentWorkspaceResponse
	if err := c.doPrivateTaskContextJSON(httpReq, &out); err != nil {
		return privateComponentWorkspaceResponse{}, err
	}
	return out.normalized(), nil
}

func (c *AIAgentPrivateTaskContextClient) getComponentDetail(ctx context.Context, teamID, componentID, bearerToken string) (privateComponentDetailResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.privateComponentDetailEndpoint(teamID, componentID), nil)
	if err != nil {
		return privateComponentDetailResponse{}, fmt.Errorf("riidoaiserver: create task context component request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+bearerToken)
	var out privateComponentDetailResponse
	if err := c.doPrivateTaskContextJSON(httpReq, &out); err != nil {
		return privateComponentDetailResponse{}, err
	}
	return out.normalized(), nil
}

func (c *AIAgentPrivateTaskContextClient) doPrivateTaskContextJSON(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("riidoaiserver: private task context request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("riidoaiserver: private task context returned HTTP %d", resp.StatusCode)
	}
	dec := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("riidoaiserver: decode private task context response: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("riidoaiserver: decode private task context response: trailing data")
	}
	return nil
}

func (c *AIAgentPrivateTaskContextClient) privateComponentWorkspaceEndpoint(componentID string) string {
	return strings.TrimRight(c.baseURL, "/") +
		"/public/components/" + url.PathEscape(componentID) +
		"/workspace"
}

func (c *AIAgentPrivateTaskContextClient) privateComponentDetailEndpoint(teamID, componentID string) string {
	return strings.TrimRight(c.baseURL, "/") +
		"/teams/" + url.PathEscape(teamID) +
		"/components/" + url.PathEscape(componentID) +
		"?getDocument=true"
}

type privateComponentWorkspaceResponse struct {
	ID            string `json:"id"`
	ComponentType string `json:"componentType"`
	Team          struct {
		ID        string `json:"id"`
		TeamKey   string `json:"teamKey"`
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
	} `json:"team"`
}

func (r privateComponentWorkspaceResponse) normalized() privateComponentWorkspaceResponse {
	r.ID = strings.TrimSpace(r.ID)
	r.ComponentType = strings.TrimSpace(r.ComponentType)
	r.Team.ID = strings.TrimSpace(r.Team.ID)
	r.Team.TeamKey = strings.TrimSpace(r.Team.TeamKey)
	r.Team.Workspace.ID = strings.TrimSpace(r.Team.Workspace.ID)
	return r
}

type privateComponentDetailResponse struct {
	ID            string                     `json:"id"`
	ComponentType string                     `json:"componentType"`
	Title         string                     `json:"title"`
	KeyNumber     string                     `json:"keyNumber"`
	Document      *privateComponentDocument  `json:"document"`
	Project       *privateComponentReference `json:"project"`
	Milestone     *privateComponentMilestone `json:"milestone"`
	Task          *privateComponentReference `json:"task"`
}

type privateComponentDocument struct {
	ID               string `json:"id"`
	TiptapDocumentID string `json:"tiptapDocumentId"`
	HTMLContent      string `json:"HTMLContent"`
}

type privateComponentReference struct {
	ID            string `json:"id"`
	ComponentType string `json:"componentType"`
	Title         string `json:"title"`
	KeyNumber     string `json:"keyNumber"`
}

type privateComponentMilestone struct {
	ID            string                     `json:"id"`
	ComponentType string                     `json:"componentType"`
	Title         string                     `json:"title"`
	KeyNumber     string                     `json:"keyNumber"`
	Project       *privateComponentReference `json:"project"`
}

func (r privateComponentDetailResponse) normalized() privateComponentDetailResponse {
	r.ID = strings.TrimSpace(r.ID)
	r.ComponentType = strings.TrimSpace(r.ComponentType)
	r.Title = strings.TrimSpace(r.Title)
	r.KeyNumber = strings.TrimSpace(r.KeyNumber)
	if r.Document != nil {
		r.Document.ID = strings.TrimSpace(r.Document.ID)
		r.Document.TiptapDocumentID = strings.TrimSpace(r.Document.TiptapDocumentID)
		r.Document.HTMLContent = strings.TrimSpace(r.Document.HTMLContent)
	}
	r.Project = normalizePrivateComponentReferencePointer(r.Project)
	r.Milestone = normalizePrivateComponentMilestonePointer(r.Milestone)
	r.Task = normalizePrivateComponentReferencePointer(r.Task)
	return r
}

func (r privateComponentDetailResponse) toAIAgentTaskContext() AIAgentTaskContext {
	var document AIAgentTaskContextDocument
	if r.Document != nil {
		document = AIAgentTaskContextDocument{
			ID:               r.Document.ID,
			TiptapDocumentID: r.Document.TiptapDocumentID,
			Content:          r.Document.HTMLContent,
			ContentFormat:    "html",
		}
	}
	var project AIAgentTaskContextReference
	if r.Project != nil {
		project = r.Project.toAIAgentTaskContextReference()
	} else if r.Milestone != nil && r.Milestone.Project != nil {
		project = r.Milestone.Project.toAIAgentTaskContextReference()
	}
	var milestone AIAgentTaskContextReference
	if r.Milestone != nil {
		milestone = r.Milestone.toAIAgentTaskContextReference()
	}
	var parentTask AIAgentTaskContextReference
	if r.Task != nil {
		parentTask = r.Task.toAIAgentTaskContextReference()
	}
	return AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{
			ID:            r.ID,
			ComponentType: r.ComponentType,
			Title:         r.Title,
			KeyNumber:     r.KeyNumber,
		},
		Document: document,
		Hierarchy: AIAgentTaskContextHierarchy{
			Project:    project,
			Milestone:  milestone,
			ParentTask: parentTask,
		},
		Repositories: []AIAgentTaskContextRepository{},
	}
}

func normalizePrivateComponentReferencePointer(ref *privateComponentReference) *privateComponentReference {
	if ref == nil {
		return nil
	}
	ref.ID = strings.TrimSpace(ref.ID)
	ref.ComponentType = strings.TrimSpace(ref.ComponentType)
	ref.Title = strings.TrimSpace(ref.Title)
	ref.KeyNumber = strings.TrimSpace(ref.KeyNumber)
	return ref
}

func normalizePrivateComponentMilestonePointer(ref *privateComponentMilestone) *privateComponentMilestone {
	if ref == nil {
		return nil
	}
	ref.ID = strings.TrimSpace(ref.ID)
	ref.ComponentType = strings.TrimSpace(ref.ComponentType)
	ref.Title = strings.TrimSpace(ref.Title)
	ref.KeyNumber = strings.TrimSpace(ref.KeyNumber)
	ref.Project = normalizePrivateComponentReferencePointer(ref.Project)
	return ref
}

func (r privateComponentReference) toAIAgentTaskContextReference() AIAgentTaskContextReference {
	return AIAgentTaskContextReference{
		ID:        r.ID,
		Title:     r.Title,
		KeyNumber: r.KeyNumber,
	}
}

func (r privateComponentMilestone) toAIAgentTaskContextReference() AIAgentTaskContextReference {
	return AIAgentTaskContextReference{
		ID:        r.ID,
		Title:     r.Title,
		KeyNumber: r.KeyNumber,
	}
}
