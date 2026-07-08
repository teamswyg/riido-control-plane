package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientEditabilityErrorBranches(t *testing.T) {
	for _, tc := range []struct {
		name   string
		token  string
		scopes []string
		store  AIAgentClientStore
		want   int
	}{
		{name: "missing auth", store: &failingEditabilityStore{}, want: http.StatusUnauthorized},
		{
			name:   "forbidden scope",
			token:  "user-token",
			scopes: []string{"component-task:task-a:read"},
			store:  &failingEditabilityStore{},
			want:   http.StatusForbidden,
		},
		{
			name:   "backend error",
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			store:  &failingEditabilityStore{err: errors.New("editability reader failed")},
			want:   http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newEditabilityErrorTestServer(t, tc.scopes, tc.store)
			req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/agents/agent-owned-codex/editability", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("editability status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
