package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestTraceHTTPRouteClassifiesCoreAgentDaemonAndComponentRoutes(t *testing.T) {
	for _, tt := range []struct {
		method string
		path   string
		want   string
	}{
		{
			method: http.MethodPost,
			path:   " /v1/agents/agent-sensitive/poll ",
			want:   "/v1/agents/{agent_id}/poll",
		},
		{
			method: http.MethodPost,
			path:   "/v1/agents/agent-sensitive/events",
			want:   "/v1/agents/{agent_id}/events",
		},
		{
			method: http.MethodPost,
			path:   "/v1/agents/agent-sensitive/thread-progress",
			want:   "/v1/agents/{agent_id}/thread-progress",
		},
		{
			method: http.MethodGet,
			path:   "/v1/agents/agent-sensitive/poll",
		},
		{
			method: http.MethodPost,
			path:   "/v1/agents/agent-sensitive/unknown",
		},
		{
			method: http.MethodPost,
			path:   "/v1/daemon/runtime-snapshot",
			want:   "/v1/daemon/runtime-snapshot",
		},
		{
			method: http.MethodGet,
			path:   "/v1/daemon/agent-bindings",
			want:   "/v1/daemon/agent-bindings",
		},
		{
			method: http.MethodPost,
			path:   "/v1/daemon/agent-bindings",
		},
		{
			method: http.MethodPost,
			path:   "/v1/component-tasks/task-sensitive",
			want:   "/v1/component-tasks/{task_id}",
		},
		{
			method: http.MethodGet,
			path:   "/v1/component-tasks/task-sensitive/events",
			want:   "/v1/component-tasks/{task_id}/events",
		},
		{
			method: http.MethodGet,
			path:   "/v1/component-tasks/task-sensitive",
		},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if got := traceHTTPRoute(tt.method, tt.path); got != tt.want {
				t.Fatalf("traceHTTPRoute() = %q, want %q", got, tt.want)
			}
		})
	}
}
