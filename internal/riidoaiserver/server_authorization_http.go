package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
)

func (s Server) authorizeAgentCatalog(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AgentCatalogPrincipal, bool) {
	result, ok := s.authorizeRequest(w, r, req)
	if !ok {
		return AgentCatalogPrincipal{}, false
	}
	principal := AgentCatalogPrincipalFromAuthorization(result)
	if err := principal.Validate(); err != nil {
		writeError(w, http.StatusForbidden, "forbidden")
		return AgentCatalogPrincipal{}, false
	}
	return principal, true
}

func (s Server) authorizeAIAgentClient(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	req.WorkspaceID = strings.TrimSpace(aiAgentWorkspaceIDFromRequest(r))
	result, ok := s.authorizeRequest(w, r, req)
	if !ok {
		return AuthorizationResult{}, false
	}
	if result.WorkspaceID == "" {
		result.WorkspaceID = req.WorkspaceID
	}
	if strings.TrimSpace(result.PrincipalID) == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return AuthorizationResult{}, false
	}
	return result, true
}

func (s Server) authorizeRequest(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	if req.Resource == AuthorizationResourceAgent {
		if result, handled := s.authorizeDeviceCredential(w, r, req); handled {
			if strings.TrimSpace(result.PrincipalID) == "" {
				return AuthorizationResult{}, false
			}
			return result, true
		}
	}
	if s.config.Authorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "scoped request authorizer is not configured")
		return AuthorizationResult{}, false
	}
	token, ok := requestToken(r)
	if !ok {
		writeUnauthorized(w)
		return AuthorizationResult{}, false
	}
	result, err := s.config.Authorizer.Authorize(r.Context(), token, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorizationForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAuthorizationUnauthenticated):
			writeUnauthorized(w)
		default:
			writeError(w, http.StatusServiceUnavailable, err.Error())
		}
		return AuthorizationResult{}, false
	}
	return result, true
}
