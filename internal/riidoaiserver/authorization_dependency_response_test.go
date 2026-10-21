package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type canaryFailureAuthorizer struct{ err error }

func (a canaryFailureAuthorizer) Authorize(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
	return AuthorizationResult{}, a.err
}

func TestAuthorizationDependencyResponseIsPrivateAndPreservesStatus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"untyped", errors.New("https://" + externalAuthCanary + ".invalid/" + externalAuthCanary), 503},
		{"typed", externalAuthorizerFailed("upstream_http", "none", 503, errors.New(externalAuthCanary)), 503},
		{"deadline", context.DeadlineExceeded, 503},
		{"cancelled", context.Canceled, 503},
		{"unauthenticated", ErrAuthorizationUnauthenticated, 401},
		{"forbidden", ErrAuthorizationForbidden, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{config: ServerConfig{Authorizer: canaryFailureAuthorizer{err: tc.err}}}
			req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
			req.Header.Set(aiAgentTokenHeader, externalAuthCanary)
			response := httptest.NewRecorder()
			if _, ok := s.authorizeRequest(response, req, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead}); ok {
				t.Fatal("failed authorization was allowed")
			}
			if response.Code != tc.status || strings.Contains(response.Body.String(), externalAuthCanary) {
				t.Fatal("dependency response leaked details or changed HTTP status")
			}
			if tc.status == 503 && !strings.Contains(response.Body.String(), "authorization dependency unavailable") {
				t.Fatal("dependency response lost fixed message")
			}
		})
	}
}
