package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failingEnrollDeviceCredentialStore struct {
	err error
}

func (s failingEnrollDeviceCredentialStore) EnrollDeviceCredential(context.Context, AuthorizationResult, string, EnrollDeviceRequest) (EnrollDeviceResponse, error) {
	return EnrollDeviceResponse{}, s.err
}

func (s failingEnrollDeviceCredentialStore) AuthorizeDeviceCredential(context.Context, string, string, AuthorizationRequest) (AuthorizationResult, error) {
	return AuthorizationResult{}, ErrAuthorizationUnauthenticated
}

func enrollErrorTestServer(t *testing.T, err error) http.Handler {
	t.Helper()
	return NewServer(ServerConfig{
		DeviceCredentials: failingEnrollDeviceCredentialStore{err: err},
		Authorizer:        aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:create"}, "user-1"),
	}).Handler()
}

func TestHTTPDesktopDeviceEnrollErrorBranches(t *testing.T) {
	tests := []struct {
		name    string
		server  http.Handler
		body    string
		want    int
		wantMsg string
	}{
		{
			name:    "missing store",
			server:  NewServer(ServerConfig{}).Handler(),
			body:    `{"display_name":"Mac"}`,
			want:    http.StatusServiceUnavailable,
			wantMsg: "device credential store is not configured",
		},
		{
			name:    "invalid json",
			server:  enrollErrorTestServer(t, nil),
			body:    `{"display_name":`,
			want:    http.StatusBadRequest,
			wantMsg: "unexpected EOF",
		},
		{
			name:    "store error",
			server:  enrollErrorTestServer(t, errors.New("enroll failed")),
			body:    `{"display_name":"Mac"}`,
			want:    http.StatusBadRequest,
			wantMsg: "enroll failed",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", strings.NewReader(tc.body))
			req.Header.Set(aiAgentTokenHeader, "ai-agent-token")
			resp := httptest.NewRecorder()
			tc.server.ServeHTTP(resp, req)
			if resp.Code != tc.want || !strings.Contains(resp.Body.String(), tc.wantMsg) {
				t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
			}
		})
	}
}
