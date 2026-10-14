package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDeviceRoutesErrorBranches(t *testing.T) {
	t.Run("store not configured", func(t *testing.T) {
		server := NewServer(ServerConfig{
			Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
		}).Handler()
		req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemon", nil)
		req.Header.Set("Authorization", "Bearer ai-agent-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusServiceUnavailable {
			t.Fatalf("device routes status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	for _, tc := range []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{"device root", http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook", http.StatusNotFound},
		{"invalid suffix", http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/unknown", http.StatusNotFound},
		{"wrong daemon action method", http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemon/restart", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer user-token")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("device route status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
