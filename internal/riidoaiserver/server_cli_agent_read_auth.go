package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
)

const (
	CLIAgentReadTokenPrefix           = "riido_cli_agent_v1_"
	AuthorizationResourceCLIAgentRead = "cli_task_agents_read"
)

// Recognize this credential before every generic/device authorization path.
// A disabled or rejected dedicated reader must never try legacy or Auth-v2.
func (s Server) authorizeCLIAgentRead(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	if !hasCLIAgentReadCredential(r) {
		return AuthorizationResult{}, false
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	token, ok := cliAgentReadRequestToken(r, req)
	if !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return AuthorizationResult{}, true
	}
	if s.config.CLIAgentReadAuthorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "CLI agent authorization is unavailable")
		return AuthorizationResult{}, true
	}
	req.Resource = AuthorizationResourceCLIAgentRead
	result, err := s.config.CLIAgentReadAuthorizer.Authorize(r.Context(), token, req)
	if err == nil && (!cliInstallationID(result.PrincipalID) || result.WorkspaceID != req.WorkspaceID) {
		err = ErrAuthorizationForbidden
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorizationForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAuthorizationUnauthenticated), errors.Is(err, ErrAuthorizationInvalidCredential):
			writeUnauthorized(w)
		default:
			writeError(w, http.StatusServiceUnavailable, "CLI agent authorization is unavailable")
		}
		return AuthorizationResult{}, true
	}
	return result, true
}

func hasCLIAgentReadCredential(r *http.Request) bool {
	for _, name := range []string{"Authorization", aiAgentTokenHeader} {
		for _, value := range r.Header.Values(name) {
			if strings.Contains(value, CLIAgentReadTokenPrefix) {
				return true
			}
		}
	}
	return false
}
