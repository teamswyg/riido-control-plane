package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientTaskThreadSubscriptionErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		path   string
		token  string
		scopes []string
		store  AIAgentClientStore
		want   int
	}{
		{"missing workspace", "/v1/client/ai-agent/tasks/task-a/thread-stream-subscription", "", nil, NewDevelopmentAIAgentClientStore(), http.StatusNotFound},
		{"missing auth", subscriptionErrorTestPath(), "", nil, NewDevelopmentAIAgentClientStore(), http.StatusUnauthorized},
		{"forbidden scope", subscriptionErrorTestPath(), "user-token", []string{"component-task:task-a:read"}, NewDevelopmentAIAgentClientStore(), http.StatusForbidden},
		{
			name:   "subscription reader error",
			path:   subscriptionErrorTestPath(),
			token:  "user-token",
			scopes: []string{"ai-agent:*"},
			store: taskThreadReadErrorStore{
				subscriptionErr: errors.New("subscription reader failed"),
			},
			want: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newTaskThreadReadErrorTestServer(t, tc.scopes, tc.store, nil)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("subscription status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}

func TestHTTPAIAgentClientTaskThreadSubscriptionReconcileError(t *testing.T) {
	assignment := NewStore()
	t.Cleanup(assignment.Close)
	server := newTaskThreadReadErrorTestServer(
		t,
		[]string{"ai-agent:*"},
		taskThreadReadErrorStore{reconcileErr: errors.New("projection reconcile failed")},
		assignment,
	)
	req := httptest.NewRequest(http.MethodGet, subscriptionErrorTestPath(), nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadGateway {
		t.Fatalf("subscription status=%d want=%d body=%s", resp.Code, http.StatusBadGateway, resp.Body.String())
	}
}

func subscriptionErrorTestPath() string {
	return "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-thread-read/thread-stream-subscription"
}
