package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientTaskAssignableAgentsErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		token  string
		scopes []string
		store  AIAgentClientStore
		want   int
	}{
		{
			name:  "missing auth",
			store: NewDevelopmentAIAgentClientStore(),
			want:  http.StatusUnauthorized,
		},
		{
			name:   "forbidden scope",
			token:  "user-token",
			scopes: []string{"component-task:task-assignable-errors:read"},
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusForbidden,
		},
		{
			name:   "assignable reader error",
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			store: assignableAgentsErrorStore{
				DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
				err:                           errors.New("assignable reader failed"),
			},
			want: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newTaskThreadReadErrorTestServer(t, tc.scopes, tc.store, nil)
			req := httptest.NewRequest(http.MethodGet, assignableAgentsErrorTestPath(), nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("assignable status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}

func TestHTTPAIAgentClientTaskAssignableAgentsReconcileError(t *testing.T) {
	assignment := NewStore()
	t.Cleanup(assignment.Close)
	server := newTaskThreadReadErrorTestServer(
		t,
		[]string{"ai-agent:*"},
		assignableAgentsErrorStore{
			DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(),
			reconcileErr:                  errors.New("projection reconcile failed"),
		},
		assignment,
	)
	req := httptest.NewRequest(http.MethodGet, assignableAgentsErrorTestPath(), nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadGateway {
		t.Fatalf("assignable status=%d want=%d body=%s", resp.Code, http.StatusBadGateway, resp.Body.String())
	}
}
