package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPComponentTaskRouteAndAssignErrorBranches(t *testing.T) {
	for _, tc := range []struct {
		name  string
		path  string
		token bool
		want  int
	}{
		{
			name: "malformed route",
			path: "/v1/component-tasks/task-a",
			want: http.StatusNotFound,
		},
		{
			name: "unsupported suffix",
			path: "/v1/component-tasks/task-a/restart",
			want: http.StatusNotFound,
		},
		{
			name: "missing assign auth",
			path: "/v1/component-tasks/task-a/assignment",
			want: http.StatusUnauthorized,
		},
		{
			name:  "assignment store error",
			path:  "/v1/component-tasks/task-a/assignment",
			token: true,
			want:  http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := NewServer(ServerConfig{
				Assignment: &handlerAssignmentStore{},
				Authorizer: assignmentHTTPAuthorizer(
					t,
					[]string{"component-task:task-a:assign"},
				),
			}).Handler()
			req := httptest.NewRequest(
				http.MethodPost,
				tc.path,
				strings.NewReader(`{"component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex","prompt":"ship it"}`),
			)
			if tc.token {
				req.Header.Set("Authorization", "Bearer assignment-token")
			}
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			if resp.Code != tc.want {
				t.Fatalf("component task status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}
