package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
)

const (
	CLIThreadReadTokenPrefix           = "riido_cli_thread_v1_"
	AuthorizationResourceCLIThreadRead = "cli_task_thread_read"
)

// Recognize this credential before every generic/device authorization path.
// A disabled or rejected dedicated reader must never try legacy or Auth-v2.
func (s Server) authorizeCLIThreadRead(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	if !hasCLIThreadReadCredential(r) {
		return AuthorizationResult{}, false
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	token, ok := cliThreadReadRequestToken(r, req)
	if !ok {
		writeError(w, http.StatusForbidden, "forbidden")
		return AuthorizationResult{}, true
	}
	if s.config.CLIThreadReadAuthorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "CLI thread authorization is unavailable")
		return AuthorizationResult{}, true
	}
	req.Resource = AuthorizationResourceCLIThreadRead
	result, err := s.config.CLIThreadReadAuthorizer.Authorize(r.Context(), token, req)
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
			writeError(w, http.StatusServiceUnavailable, "CLI thread authorization is unavailable")
		}
		return AuthorizationResult{}, true
	}
	return result, true
}

func hasCLIThreadReadCredential(r *http.Request) bool {
	for _, name := range []string{"Authorization", aiAgentTokenHeader} {
		for _, value := range r.Header.Values(name) {
			if strings.Contains(value, CLIThreadReadTokenPrefix) {
				return true
			}
		}
	}
	return false
}
