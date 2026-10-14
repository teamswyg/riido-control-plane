package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestTraceHTTPRouteClassifiesV1AIAgentClientRoutes(t *testing.T) {
	tests := map[string]string{
		"/v1/client/ai-agent/bootstrap":                         "/v1/client/ai-agent/bootstrap",
		"/v1/client/ai-agent/devices":                           "/v1/client/ai-agent/devices",
		"/v1/client/ai-agent/devices/dev-sensitive/daemon":      "/v1/client/ai-agent/devices/{device_id}/daemon",
		"/v1/client/ai-agent/devices/dev-sensitive/daemons":     "/v1/client/ai-agent/devices/{device_id}/daemons",
		"/v1/client/ai-agent/onboarding/fixtures":               "/v1/client/ai-agent/onboarding/fixtures",
		"/v1/client/ai-agent/profile-thumbnails/uploads":        "/v1/client/ai-agent/profile-thumbnails/uploads",
		"/v1/client/ai-agent/tasks/task-sensitive/threads":      "/v1/client/ai-agent/tasks/{task_id}/threads",
		"/v1/client/ai-agent/tasks/task-sensitive/comments":     "/v1/client/ai-agent/tasks/{task_id}/comments",
		"/v1/client/ai-agent/threads/thread-sensitive/messages": "/v1/client/ai-agent/threads/{thread_id}/messages",
		"/v1/client/ai-agent/agents/agent-sensitive/daemon":     "/v1/client/ai-agent/agents/{agent_id}/daemon",
	}
	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			if got := traceHTTPRoute(http.MethodGet, path); got != want {
				t.Fatalf("traceHTTPRoute(%q) = %q, want %q", path, got, want)
			}
		})
	}
}

func TestAIAgentClientV1RouteRejectsNonAIAgentClientPaths(t *testing.T) {
	tests := []string{
		"/v1/client/ai-agent",
		"/v1/client/ai-agentish/bootstrap",
		"/v1/client/ai-agent/unknown",
		"/v1/client/ai-agent/events/extra",
		"/v1/client/ai-agent/onboarding",
		"/v1/client/ai-agent/profile-thumbnails",
	}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			if got := aiAgentClientV1Route(path); got != "" {
				t.Fatalf("aiAgentClientV1Route(%q) = %q, want empty", path, got)
			}
		})
	}
}
