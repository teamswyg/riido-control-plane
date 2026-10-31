package authpep

import (
	"encoding/json"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func (a *CLIThreadReadAuthorizer) decode(raw []byte, req AuthorizationRequest) (AuthorizationResult, error) {
	if !utf8.Valid(raw) || rejectDuplicateTopLevelJWTKeys(raw) != nil {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || len(fields) != 10 {
		return AuthorizationResult{}, errCLIThreadReadUnavailable
	}
	for _, key := range []string{"schema_version", "allowed", "principal_id", "roles", "resource", "purpose", "workspace_id", "task_id", "expires_at", "checked_at"} {
		value, found := fields[key]
		if !found || string(value) == "null" {
			return AuthorizationResult{}, errCLIThreadReadUnavailable
		}
	}
	var result cliThreadReadResponse
	if json.Unmarshal(raw, &result) != nil || result.Schema != cliThreadReadSchema || !result.Allowed || !cliThreadReadID(result.PrincipalID) || result.Resource != a.resource || result.Purpose != cliThreadReadPurpose || result.WorkspaceID != req.WorkspaceID || result.TaskID != req.TaskID {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	now := a.now().UTC()
	if result.CheckedAt.Before(now.Add(-5*time.Second)) || result.CheckedAt.After(now.Add(5*time.Second)) || !result.ExpiresAt.After(now) || !result.ExpiresAt.After(result.CheckedAt) || result.ExpiresAt.Sub(result.CheckedAt) > time.Minute {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	roles := []riidoaiserver.AgentCatalogRole{}
	if len(result.Roles) == 1 && result.Roles[0] == "admin" {
		roles = append(roles, riidoaiserver.AgentCatalogRoleAdmin)
	} else if len(result.Roles) != 0 {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}
	return AuthorizationResult{PrincipalID: result.PrincipalID, WorkspaceID: result.WorkspaceID, Roles: roles}, nil
}
