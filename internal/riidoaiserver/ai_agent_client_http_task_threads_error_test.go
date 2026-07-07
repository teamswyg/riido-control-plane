package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientTaskThreadsErrors(t *testing.T) {
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
			scopes: []string{"component-task:task-a:read"},
			store:  NewDevelopmentAIAgentClientStore(),
			want:   http.StatusForbidden,
		},
		{
			name:   "thread reader error",
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			store: taskThreadReadErrorStore{
				listErr: errors.New("thread reader failed"),
			},
			want: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newTaskThreadReadErrorTestServer(t, tc.scopes, tc.store, nil)
			req := httptest.NewRequest(http.MethodGet, taskThreadsErrorTestPath(), nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("threads status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}

func TestHTTPAIAgentClientTaskThreadsReconcileError(t *testing.T) {
	assignment := NewStore()
	t.Cleanup(assignment.Close)
	server := newTaskThreadReadErrorTestServer(
		t,
		[]string{"ai-agent:*"},
		taskThreadReadErrorStore{reconcileErr: errors.New("projection reconcile failed")},
		assignment,
	)
	req := httptest.NewRequest(http.MethodGet, taskThreadsErrorTestPath(), nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadGateway {
		t.Fatalf("threads status=%d want=%d body=%s", resp.Code, http.StatusBadGateway, resp.Body.String())
	}
}

func taskThreadsErrorTestPath() string {
	return "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-thread-read/threads"
}
