package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientAssignTaskRejectsBoundaryErrors(t *testing.T) {
	server, store := newAssignTaskBoundaryServer(t, true)
	path := "/v1/client/ai-agent/tasks/task-boundary/assignment"
	cases := []struct {
		name  string
		token string
		body  string
		want  int
	}{
		{name: "missing auth", body: `{"agent_id":"agent-owned-codex"}`, want: http.StatusUnauthorized},
		{name: "malformed json", token: "user-token", body: `{`, want: http.StatusBadRequest},
		{name: "missing agent", token: "user-token", body: `{}`, want: http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := serveAssignTaskBoundary(server, path, tc.token, tc.body)
			if resp.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", resp.Code, tc.want, resp.Body.String())
			}
			assertAssignTaskBoundaryPollNone(t, store)
		})
	}
}

func TestHTTPAIAgentClientAssignTaskRequiresAssignmentStore(t *testing.T) {
	server, _ := newAssignTaskBoundaryServer(t, false)
	resp := serveAssignTaskBoundary(server, "/v1/client/ai-agent/tasks/task-boundary/assignment", "", `{}`)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d body=%s", resp.Code, http.StatusServiceUnavailable, resp.Body.String())
	}
}
