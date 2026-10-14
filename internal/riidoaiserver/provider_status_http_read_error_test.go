package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPProviderStatusReadFailureContracts(t *testing.T) {
	cases := []struct {
		name   string
		server http.Handler
		token  string
		want   int
	}{
		{
			name:   "missing reader",
			server: NewServer(ServerConfig{}).Handler(),
			want:   http.StatusServiceUnavailable,
		},
		{
			name: "forbidden scope",
			server: NewServer(ServerConfig{
				ProviderStatus: newProviderStatusTestStore(time.Now().UTC()),
				Authorizer:     providerStatusAuthorizer(t, "agent:agent-b:provider-status:read"),
			}).Handler(),
			token: "agent-token",
			want:  http.StatusForbidden,
		},
		{
			name: "not found",
			server: NewServer(ServerConfig{
				ProviderStatus: newProviderStatusTestStore(time.Now().UTC()),
				Authorizer:     providerStatusAuthorizer(t, "agent:agent-a:provider-status:read"),
			}).Handler(),
			token: "agent-token",
			want:  http.StatusNotFound,
		},
		{
			name: "reader error",
			server: NewServer(ServerConfig{
				ProviderRead: providerStatusFailingReader{},
				Authorizer:   providerStatusAuthorizer(t, "agent:agent-a:provider-status:read"),
			}).Handler(),
			token: "agent-token",
			want:  http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/agents/agent-a/provider-status", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			tc.server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("status=%d body=%s want=%d", resp.Code, resp.Body.String(), tc.want)
			}
		})
	}
}
