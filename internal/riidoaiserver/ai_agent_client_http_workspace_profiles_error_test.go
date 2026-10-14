package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientWorkspaceAssignedAgentProfilesErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		token  string
		scopes []string
		store  AIAgentClientStore
		want   int
	}{
		{
			name:   "unconfigured store",
			method: http.MethodGet,
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			want:   http.StatusServiceUnavailable,
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "missing auth",
			method: http.MethodGet,
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusUnauthorized,
		},
		{
			name:   "forbidden scope",
			method: http.MethodGet,
			token:  "user-token",
			scopes: []string{"component-task:task-a:read"},
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusForbidden,
		},
		{
			name:   "assigned profile reader error",
			method: http.MethodGet,
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			store: workspaceProfilesErrorStore{
				err: errors.New("assigned profile reader failed"),
			},
			want: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newWorkspaceProfilesErrorTestServer(t, tc.scopes, tc.store, nil)
			req := httptest.NewRequest(tc.method, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/assigned-agent-profiles", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("assigned profiles status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
