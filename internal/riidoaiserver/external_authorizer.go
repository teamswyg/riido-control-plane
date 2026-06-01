package riidoaiserver

import (
	"bytes"
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
	ExternalAuthorizerRequestSchemaVersion  = "riido-external-authorizer-request.v1"
	ExternalAuthorizerResponseSchemaVersion = "riido-external-authorizer-response.v1"
	DefaultExternalHTTPAuthorizerTimeout    = 2 * time.Second
)

type ExternalHTTPAuthorizerConfig struct {
	Endpoint   string
	Audience   string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type ExternalHTTPAuthorizer struct {
	endpoint   string
	audience   string
	timeout    time.Duration
	httpClient *http.Client
}

type externalAuthorizerRequest struct {
	SchemaVersion string                       `json:"schema_version"`
	BearerToken   string                       `json:"bearer_token"`
	Audience      string                       `json:"audience,omitempty"`
	Request       externalAuthorizationRequest `json:"request"`
}

type externalAuthorizationRequest struct {
	Resource    AuthorizationResource `json:"resource"`
	Action      AuthorizationAction   `json:"action"`
	WorkspaceID string                `json:"workspace_id,omitempty"`
	AgentID     string                `json:"agent_id,omitempty"`
	TaskID      string                `json:"task_id,omitempty"`
}

type externalAuthorizerResponse struct {
	SchemaVersion string             `json:"schema_version"`
	Allowed       bool               `json:"allowed"`
	PrincipalID   string             `json:"principal_id,omitempty"`
	Roles         []AgentCatalogRole `json:"roles,omitempty"`
}

func NewExternalHTTPAuthorizer(config ExternalHTTPAuthorizerConfig) (*ExternalHTTPAuthorizer, error) {
	endpoint, err := normalizeExternalAuthorizerEndpoint(strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultExternalHTTPAuthorizerTimeout
	}
	if timeout < 0 {
		return nil, errors.New("riidoaiserver: external authorizer timeout must be positive")
	}
	return &ExternalHTTPAuthorizer{
		endpoint:   endpoint,
		audience:   strings.TrimSpace(config.Audience),
		timeout:    timeout,
		httpClient: externalAuthorizerHTTPClient(config.HTTPClient),
	}, nil
}

func (a *ExternalHTTPAuthorizer) Authorize(ctx context.Context, bearerToken string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	if a == nil {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	bearerToken = strings.TrimSpace(bearerToken)
	if bearerToken == "" {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	payload, err := json.Marshal(externalAuthorizerRequest{
		SchemaVersion: ExternalAuthorizerRequestSchemaVersion,
		BearerToken:   bearerToken,
		Audience:      a.audience,
		Request: externalAuthorizationRequest{
			Resource:    req.Resource,
			Action:      req.Action,
			WorkspaceID: strings.TrimSpace(req.WorkspaceID),
			AgentID:     strings.TrimSpace(req.AgentID),
			TaskID:      strings.TrimSpace(req.TaskID),
		},
	})
	if err != nil {
		return AuthorizationResult{}, fmt.Errorf("riidoaiserver: encode external authorizer request: %w", err)
	}
	requestCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodPost, a.endpoint, bytes.NewReader(payload))
	if err != nil {
		return AuthorizationResult{}, fmt.Errorf("riidoaiserver: create external authorizer request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return AuthorizationResult{}, fmt.Errorf("riidoaiserver: external authorizer request failed: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	case http.StatusForbidden:
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return AuthorizationResult{}, fmt.Errorf("riidoaiserver: external authorizer returned HTTP %d", resp.StatusCode)
	}

	var out externalAuthorizerResponse
	dec := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return AuthorizationResult{}, fmt.Errorf("riidoaiserver: decode external authorizer response: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return AuthorizationResult{}, errors.New("riidoaiserver: decode external authorizer response: trailing data")
	}
	if out.SchemaVersion != ExternalAuthorizerResponseSchemaVersion {
		return AuthorizationResult{}, fmt.Errorf("riidoaiserver: unsupported external authorizer schema_version %q", out.SchemaVersion)
	}
	if !out.Allowed {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	out.PrincipalID = strings.TrimSpace(out.PrincipalID)
	if out.PrincipalID == "" {
		return AuthorizationResult{}, errors.New("riidoaiserver: external authorizer principal_id is required when allowed")
	}
	roles, err := normalizeAgentCatalogRoles(out.Roles)
	if err != nil {
		return AuthorizationResult{}, fmt.Errorf("riidoaiserver: external authorizer roles: %w", err)
	}
	return AuthorizationResult{PrincipalID: out.PrincipalID, WorkspaceID: strings.TrimSpace(req.WorkspaceID), Roles: roles}, nil
}

func normalizeExternalAuthorizerEndpoint(endpoint string) (string, error) {
	if endpoint == "" {
		return "", errors.New("riidoaiserver: external authorizer endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse external authorizer endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("riidoaiserver: external authorizer endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("riidoaiserver: external authorizer endpoint host is required")
	}
	if parsed.User != nil {
		return "", errors.New("riidoaiserver: external authorizer endpoint must not include userinfo")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("riidoaiserver: external authorizer endpoint must not include query or fragment")
	}
	return parsed.String(), nil
}

func externalAuthorizerHTTPClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: DefaultExternalHTTPAuthorizerTimeout}
}
