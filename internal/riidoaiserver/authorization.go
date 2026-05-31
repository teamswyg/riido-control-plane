package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type AuthorizationResource string

const (
	AuthorizationResourceAIAgentClient       AuthorizationResource = "ai_agent_client"
	AuthorizationResourceAgent               AuthorizationResource = "agent"
	AuthorizationResourceAgentCatalog        AuthorizationResource = "agent_catalog"
	AuthorizationResourceComponentTask       AuthorizationResource = "component_task"
	AuthorizationResourceComponentTaskEvents AuthorizationResource = "component_task_events"
	AuthorizationResourceMetrics             AuthorizationResource = "metrics"
)

type AuthorizationAction string

const (
	AuthorizationActionAssign              AuthorizationAction = "assign"
	AuthorizationActionCreate              AuthorizationAction = "create"
	AuthorizationActionDelete              AuthorizationAction = "delete"
	AuthorizationActionDeviceRead          AuthorizationAction = "device:read"
	AuthorizationActionEventsRead          AuthorizationAction = "events:read"
	AuthorizationActionEventsWrite         AuthorizationAction = "events:write"
	AuthorizationActionHeartbeat           AuthorizationAction = "heartbeat"
	AuthorizationActionPoll                AuthorizationAction = "poll"
	AuthorizationActionProviderStatusRead  AuthorizationAction = "provider-status:read"
	AuthorizationActionProviderStatusWrite AuthorizationAction = "provider-status:write"
	AuthorizationActionRead                AuthorizationAction = "read"
	AuthorizationActionStream              AuthorizationAction = "stream"
	AuthorizationActionStop                AuthorizationAction = "stop"
	AuthorizationActionUpdate              AuthorizationAction = "update"
)

var (
	ErrAuthorizationUnauthenticated = errors.New("riidoaiserver: unauthenticated")
	ErrAuthorizationForbidden       = errors.New("riidoaiserver: forbidden")
)

type AuthorizationRequest struct {
	Resource AuthorizationResource
	Action   AuthorizationAction
	AgentID  string
	TaskID   string
}

type AuthorizationResult struct {
	PrincipalID string
	Roles       []AgentCatalogRole
}

type RequestAuthorizer interface {
	Authorize(ctx context.Context, bearerToken string, req AuthorizationRequest) (AuthorizationResult, error)
}

type FallbackAuthorizer struct {
	authorizers []RequestAuthorizer
}

func NewFallbackAuthorizer(authorizers ...RequestAuthorizer) (*FallbackAuthorizer, error) {
	filtered := make([]RequestAuthorizer, 0, len(authorizers))
	for _, authorizer := range authorizers {
		if authorizer != nil {
			filtered = append(filtered, authorizer)
		}
	}
	if len(filtered) == 0 {
		return nil, errors.New("request authorizer is required")
	}
	return &FallbackAuthorizer{authorizers: filtered}, nil
}

func (a *FallbackAuthorizer) Authorize(ctx context.Context, bearerToken string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	if a == nil || len(a.authorizers) == 0 {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	for _, authorizer := range a.authorizers {
		result, err := authorizer.Authorize(ctx, bearerToken, req)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, ErrAuthorizationUnauthenticated) {
			continue
		}
		return AuthorizationResult{}, err
	}
	return AuthorizationResult{}, ErrAuthorizationUnauthenticated
}

type StaticTokenCredential struct {
	PrincipalID string             `json:"principal_id"`
	Token       string             `json:"token,omitempty"`
	TokenSHA256 string             `json:"token_sha256,omitempty"`
	Scopes      []string           `json:"scopes"`
	Roles       []AgentCatalogRole `json:"roles,omitempty"`
}

type StaticTokenAuthorizer struct {
	credentials []staticTokenCredential
}

type staticTokenCredential struct {
	principalID string
	tokenHash   [sha256.Size]byte
	scopes      []string
	roles       []AgentCatalogRole
}

func NewStaticTokenAuthorizer(credentials []StaticTokenCredential) (*StaticTokenAuthorizer, error) {
	compiled := make([]staticTokenCredential, 0, len(credentials))
	seenHashes := map[string]struct{}{}
	for i, credential := range credentials {
		credential.PrincipalID = strings.TrimSpace(credential.PrincipalID)
		credential.Token = strings.TrimSpace(credential.Token)
		credential.TokenSHA256 = strings.ToLower(strings.TrimSpace(credential.TokenSHA256))
		if credential.PrincipalID == "" {
			return nil, fmt.Errorf("authz token %d: principal_id is required", i)
		}
		tokenSet := credential.Token != ""
		hashSet := credential.TokenSHA256 != ""
		if tokenSet == hashSet {
			return nil, fmt.Errorf("authz token %s: exactly one of token or token_sha256 is required", credential.PrincipalID)
		}
		tokenHash, err := staticCredentialHash(credential)
		if err != nil {
			return nil, fmt.Errorf("authz token %s: %w", credential.PrincipalID, err)
		}
		hashKey := hex.EncodeToString(tokenHash[:])
		if _, exists := seenHashes[hashKey]; exists {
			return nil, fmt.Errorf("authz token %s: duplicate token", credential.PrincipalID)
		}
		seenHashes[hashKey] = struct{}{}
		scopes, err := normalizeAuthorizationScopes(credential.Scopes)
		if err != nil {
			return nil, fmt.Errorf("authz token %s: %w", credential.PrincipalID, err)
		}
		roles, err := normalizeAgentCatalogRoles(credential.Roles)
		if err != nil {
			return nil, fmt.Errorf("authz token %s: %w", credential.PrincipalID, err)
		}
		compiled = append(compiled, staticTokenCredential{
			principalID: credential.PrincipalID,
			tokenHash:   tokenHash,
			scopes:      scopes,
			roles:       roles,
		})
	}
	if len(compiled) == 0 {
		return nil, errors.New("authz token credentials are required")
	}
	return &StaticTokenAuthorizer{credentials: compiled}, nil
}

func (a *StaticTokenAuthorizer) Authorize(ctx context.Context, bearerToken string, req AuthorizationRequest) (AuthorizationResult, error) {
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
	tokenHash := sha256.Sum256([]byte(bearerToken))
	for _, credential := range a.credentials {
		if subtle.ConstantTimeCompare(tokenHash[:], credential.tokenHash[:]) != 1 {
			continue
		}
		if !authorizationScopesPermit(credential.scopes, req) {
			return AuthorizationResult{}, ErrAuthorizationForbidden
		}
		return AuthorizationResult{PrincipalID: credential.principalID, Roles: append([]AgentCatalogRole(nil), credential.roles...)}, nil
	}
	return AuthorizationResult{}, ErrAuthorizationUnauthenticated
}

func staticCredentialHash(credential StaticTokenCredential) ([sha256.Size]byte, error) {
	if credential.Token != "" {
		return sha256.Sum256([]byte(credential.Token)), nil
	}
	decoded, err := hex.DecodeString(credential.TokenSHA256)
	if err != nil || len(decoded) != sha256.Size {
		return [sha256.Size]byte{}, errors.New("token_sha256 must be 64 hex characters")
	}
	var out [sha256.Size]byte
	copy(out[:], decoded)
	return out, nil
}

func normalizeAuthorizationScopes(scopes []string) ([]string, error) {
	normalized := make([]string, 0, len(scopes))
	seen := map[string]struct{}{}
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			return nil, errors.New("scope must not be empty")
		}
		if _, exists := seen[scope]; exists {
			return nil, fmt.Errorf("duplicate scope %s", scope)
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	if len(normalized) == 0 {
		return nil, errors.New("at least one scope is required")
	}
	return normalized, nil
}

func authorizationScopesPermit(scopes []string, req AuthorizationRequest) bool {
	allowed := map[string]struct{}{}
	for _, scope := range scopes {
		allowed[scope] = struct{}{}
	}
	for _, candidate := range authorizationScopeCandidates(req) {
		if _, ok := allowed[candidate]; ok {
			return true
		}
	}
	return false
}

func authorizationScopeCandidates(req AuthorizationRequest) []string {
	switch req.Resource {
	case AuthorizationResourceAIAgentClient:
		candidates := []string{"riido:*", "ai-agent:*"}
		switch req.Action {
		case AuthorizationActionDeviceRead:
			candidates = append(candidates, "ai-agent:device:read", "ai-agent:read")
		case AuthorizationActionStream:
			candidates = append(candidates, "ai-agent:stream")
		case AuthorizationActionRead:
			candidates = append(candidates, "ai-agent:read")
			if req.AgentID != "" {
				candidates = append(candidates, "ai-agent:"+req.AgentID+":read")
			}
			if req.TaskID != "" {
				candidates = append(candidates, "task:"+req.TaskID+":read")
			}
		case AuthorizationActionCreate:
			candidates = append(candidates, "ai-agent:write", "ai-agent:create")
			if req.TaskID != "" {
				candidates = append(candidates, "task:"+req.TaskID+":comment")
			}
		case AuthorizationActionAssign:
			candidates = append(candidates, "ai-agent:write")
			if req.TaskID != "" {
				candidates = append(candidates, "task:"+req.TaskID+":assign", "task:"+req.TaskID+":write")
			}
		case AuthorizationActionStop:
			candidates = append(candidates, "ai-agent:write")
			if req.TaskID != "" {
				candidates = append(candidates, "task:"+req.TaskID+":stop", "task:"+req.TaskID+":write")
			}
		case AuthorizationActionUpdate:
			candidates = append(candidates, "ai-agent:write")
			if req.AgentID != "" {
				candidates = append(candidates, "ai-agent:"+req.AgentID+":update")
			}
		case AuthorizationActionDelete:
			candidates = append(candidates, "ai-agent:write")
			if req.AgentID != "" {
				candidates = append(candidates, "ai-agent:"+req.AgentID+":delete")
			}
		}
		return candidates
	case AuthorizationResourceMetrics:
		return []string{"riido:*", "metrics:*", "metrics:read"}
	case AuthorizationResourceComponentTask:
		return []string{
			"riido:*",
			"component-task:*",
			"component-task:*:" + string(req.Action),
			"component-task:" + req.TaskID + ":*",
			"component-task:" + req.TaskID + ":" + string(req.Action),
		}
	case AuthorizationResourceComponentTaskEvents:
		return []string{
			"riido:*",
			"component-task:*",
			"component-task:*:events:*",
			"component-task:*:" + string(req.Action),
			"component-task:" + req.TaskID + ":*",
			"component-task:" + req.TaskID + ":events:*",
			"component-task:" + req.TaskID + ":" + string(req.Action),
		}
	case AuthorizationResourceAgent:
		return []string{
			"riido:*",
			"agent:*",
			"agent:*:" + string(req.Action),
			"agent:" + req.AgentID + ":*",
			"agent:" + req.AgentID + ":" + string(req.Action),
		}
	case AuthorizationResourceAgentCatalog:
		if req.Action == AuthorizationActionCreate {
			return []string{
				"riido:*",
				"agent-catalog:*",
				"agent-catalog:write",
				"agent-catalog:create",
			}
		}
		if req.AgentID == "" {
			return []string{
				"riido:*",
				"agent-catalog:*",
				"agent-catalog:" + string(req.Action),
			}
		}
		candidates := []string{
			"riido:*",
			"agent-catalog:*",
			"agent-catalog:*:" + string(req.Action),
			"agent-catalog:" + req.AgentID + ":*",
			"agent-catalog:" + req.AgentID + ":" + string(req.Action),
		}
		if req.Action == AuthorizationActionUpdate || req.Action == AuthorizationActionDelete {
			candidates = append(candidates, "agent-catalog:write")
		}
		if req.Action == AuthorizationActionRead {
			candidates = append(candidates, "agent-catalog:read")
		}
		return candidates
	default:
		return []string{"riido:*"}
	}
}
