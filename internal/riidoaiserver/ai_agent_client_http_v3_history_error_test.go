package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientV3ThreadHistoryErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		path   string
		token  string
		scopes []string
		store  AIAgentClientStore
		want   int
	}{
		{"unconfigured store", http.MethodGet, v3HistoryErrorTestPath(), "user-token", nil, nil, http.StatusServiceUnavailable},
		{"malformed task route", http.MethodGet, "/v3/client/workspaces/workspace-dev-riid/ai-agent/tasks", "user-token", nil, NewDevelopmentAIAgentClientStore(), http.StatusNotFound},
		{"method not allowed", http.MethodPost, v3HistoryErrorTestPath(), "user-token", nil, NewDevelopmentAIAgentClientStore(), http.StatusMethodNotAllowed},
		{"missing auth", http.MethodGet, v3HistoryErrorTestPath(), "", nil, NewDevelopmentAIAgentClientStore(), http.StatusUnauthorized},
		{"forbidden scope", http.MethodGet, v3HistoryErrorTestPath(), "user-token", []string{"component-task:task-a:read"}, NewDevelopmentAIAgentClientStore(), http.StatusForbidden},
		{
			name:   "history reader error",
			method: http.MethodGet,
			path:   v3HistoryErrorTestPath(),
			token:  "user-token",
			store: v3HistoryErrorStore{
				historyErr: errors.New("history reader failed"),
			},
			want: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newTaskThreadReadErrorTestServer(t, tc.scopes, tc.store, nil)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("v3 history status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}

func TestHTTPAIAgentClientV3ThreadHistoryReconcileError(t *testing.T) {
	assignment := NewStore()
	t.Cleanup(assignment.Close)
	server := newTaskThreadReadErrorTestServer(
		t,
		[]string{"ai-agent:*"},
		v3HistoryErrorStore{reconcileErr: errors.New("projection reconcile failed")},
		assignment,
	)
	req := httptest.NewRequest(http.MethodGet, v3HistoryErrorTestPath(), nil)
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadGateway {
		t.Fatalf("v3 history status=%d want=%d body=%s", resp.Code, http.StatusBadGateway, resp.Body.String())
	}
}
