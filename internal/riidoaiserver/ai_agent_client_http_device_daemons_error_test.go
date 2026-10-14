package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDeviceDaemonsErrorBranches(t *testing.T) {
	for _, tc := range []struct {
		name   string
		token  string
		scopes []string
		store  AIAgentClientStore
		want   int
	}{
		{
			name:  "missing auth",
			store: &failingDeviceDaemonsStore{},
			want:  http.StatusUnauthorized,
		},
		{
			name:   "forbidden scope",
			token:  "user-token",
			scopes: []string{"component-task:task-a:read"},
			store:  &failingDeviceDaemonsStore{},
			want:   http.StatusForbidden,
		},
		{
			name:   "backend error",
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			store:  &failingDeviceDaemonsStore{err: errors.New("device daemon reader failed")},
			want:   http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newDeviceDaemonsErrorTestServer(t, tc.scopes, tc.store)
			req := httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-alpha/ai-agent/devices/device-dev-macbook/daemons", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("device daemons status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
