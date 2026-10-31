package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLIThreadReadRejectsOtherPathsAndAmbiguousCredentials(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*http.Request, *AuthorizationRequest)
	}{
		{"write", func(r *http.Request, q *AuthorizationRequest) {
			r.Method = http.MethodPost
			q.Action = AuthorizationActionCreate
		}},
		{"stream", func(r *http.Request, q *AuthorizationRequest) { r.URL.Path += "/stream-subscription" }},
		{"devices", func(r *http.Request, q *AuthorizationRequest) {
			r.URL.Path = "/v1/client/ai-agent/devices"
			q.TaskID = ""
		}},
		{"v1", func(r *http.Request, _ *AuthorizationRequest) {
			r.URL.Path = "/v1/client/ai-agent/tasks/task-fixture/threads"
		}},
		{"other task", func(_ *http.Request, q *AuthorizationRequest) { q.TaskID = "other" }},
		{"workspace", func(_ *http.Request, q *AuthorizationRequest) { q.WorkspaceID = "other" }},
		{"query", func(r *http.Request, _ *AuthorizationRequest) {
			r.URL.RawQuery = "cursor=x"
			r.RequestURI += "?cursor=x"
		}},
		{"duplicate", func(r *http.Request, _ *AuthorizationRequest) { r.Header.Add("Authorization", "Bearer other") }},
		{"agent header", func(r *http.Request, _ *AuthorizationRequest) { r.Header.Set(aiAgentTokenHeader, "ordinary") }},
		{"device header", func(r *http.Request, _ *AuthorizationRequest) { r.Header.Set(deviceIDHeader, "device") }},
		{"malformed", func(r *http.Request, _ *AuthorizationRequest) {
			r.Header.Set("Authorization", "Bearer "+CLIThreadReadTokenPrefix+"invalid")
		}},
		{"encoded", func(r *http.Request, _ *AuthorizationRequest) {
			r.URL.RawPath = "/v3/client/workspaces/%77orkspace-fixture/ai-agent/tasks/task-fixture/threads"
		}},
		{"trailing slash", func(r *http.Request, _ *AuthorizationRequest) { r.RequestURI += "/" }},
		{"origin", func(r *http.Request, _ *AuthorizationRequest) { r.Header.Set("Origin", "https://app.example.test") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			a := authorizationFunc(func(context.Context, string, AuthorizationRequest) (AuthorizationResult, error) {
				calls++
				return AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-fixture"}, nil
			})
			s := NewServer(ServerConfig{Authorizer: a, CLIThreadReadAuthorizer: a})
			r := httptest.NewRequest(http.MethodGet, "/v3/client/workspaces/workspace-fixture/ai-agent/tasks/task-fixture/threads", nil)
			r.Header.Set("Authorization", "Bearer "+CLIThreadReadTokenPrefix+strings.Repeat("A", 43))
			r = requestWithAIAgentWorkspaceIDAndPath(r, "workspace-fixture", "/v3/client/ai-agent/tasks/task-fixture/threads")
			q := AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, WorkspaceID: "workspace-fixture", TaskID: "task-fixture"}
			tc.mutate(r, &q)
			w := httptest.NewRecorder()
			_, ok := s.authorizeRequest(w, r, q)
			if ok || w.Code != 403 || calls != 0 {
				t.Fatal("unsupported request reached authorizer", w.Code, calls)
			}
		})
	}
}
