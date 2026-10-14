package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestTraceHTTPRouteClassifiesV2AIAgentClientRoutes(t *testing.T) {
	tests := map[string]string{
		"/v2/client/workspaces/ws-sensitive/ai-agent/bootstrap":                                       "/v2/client/workspaces/{workspace_id}/ai-agent/bootstrap",
		"/v2/client/workspaces/ws-sensitive/ai-agent/devices":                                         "/v2/client/workspaces/{workspace_id}/ai-agent/devices",
		"/v2/client/workspaces/ws-sensitive/ai-agent/devices/dev-sensitive/daemon":                    "/v2/client/workspaces/{workspace_id}/ai-agent/devices/{device_id}/daemon",
		"/v2/client/workspaces/ws-sensitive/ai-agent/devices/dev-sensitive/daemons":                   "/v2/client/workspaces/{workspace_id}/ai-agent/devices/{device_id}/daemons",
		"/v2/client/workspaces/ws-sensitive/ai-agent/onboarding/fixtures":                             "/v2/client/workspaces/{workspace_id}/ai-agent/onboarding/fixtures",
		"/v2/client/workspaces/ws-sensitive/ai-agent/profile-thumbnails/uploads":                      "/v2/client/workspaces/{workspace_id}/ai-agent/profile-thumbnails/uploads",
		"/v2/client/workspaces/ws-sensitive/ai-agent/tasks/task-sensitive/assignable-agents":          "/v2/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/assignable-agents",
		"/v2/client/workspaces/ws-sensitive/ai-agent/tasks/task-sensitive/agent-assignments":          "/v2/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/agent-assignments",
		"/v2/client/workspaces/ws-sensitive/ai-agent/tasks/task-sensitive/thread-stream-subscription": "/v2/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/thread-stream-subscription",
		"/v2/client/workspaces/ws-sensitive/ai-agent/tasks/task-sensitive/threads/thread-a/messages":  "/v2/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads/{thread_id}/messages",
		"/v3/client/workspaces/ws-sensitive/ai-agent/tasks/task-sensitive/threads":                    "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads",
		"/v2/client/workspaces/ws-sensitive/ai-agent/threads/thread-sensitive/messages":               "/v2/client/workspaces/{workspace_id}/ai-agent/threads/{thread_id}/messages",
		"/v2/client/workspaces/ws-sensitive/ai-agent/agent-assignments/agent-sensitive/stop":          "/v2/client/workspaces/{workspace_id}/ai-agent/agent-assignments/{agent_id}/stop",
		"/v2/client/workspaces/ws-sensitive/ai-agent/agents/agent-sensitive/daemon/restart":           "/v2/client/workspaces/{workspace_id}/ai-agent/agents/{agent_id}/daemon/{action}",
	}
	for path, want := range tests {
		if got := traceHTTPRoute(http.MethodGet, path); got != want {
			t.Fatalf("traceHTTPRoute(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestAIAgentTaskThreadMessageRouteKeepsMessageVocabulary(t *testing.T) {
	taskBase := "/v2/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}"
	segments := []string{"task-sensitive", "threads", "thread-sensitive", "messages"}
	got := aiAgentClientTaskThreadMessageRoute(taskBase, segments)
	want := taskBase + "/threads/{thread_id}/messages"
	if got != want {
		t.Fatalf("aiAgentClientTaskThreadMessageRoute = %q, want %q", got, want)
	}
}

func TestHTTPMetricRouteUsesV2AIAgentRouteVocabulary(t *testing.T) {
	path := "/v2/client/workspaces/ws-sensitive/ai-agent/bootstrap"
	got := httpMetricRoute(http.MethodGet, path, "/v2/client/workspaces/", http.StatusOK)
	want := "/v2/client/workspaces/{workspace_id}/ai-agent/bootstrap"
	if got != want {
		t.Fatalf("httpMetricRoute = %q, want %q", got, want)
	}
}
