package riidoaiserver

import "testing"

func TestAIAgentClientTaskHTTPRouteClassifiesTaskScopedRoutes(t *testing.T) {
	base := "/v2/client/workspaces/{workspace_id}/ai-agent"
	taskBase := base + "/tasks/{task_id}"
	tests := []struct {
		name     string
		segments []string
		want     string
	}{
		{"assigned profiles", []string{"assigned-agent-profiles"}, base + "/tasks/assigned-agent-profiles"},
		{"assignable", []string{"task-a", "assignable-agents"}, taskBase + "/assignable-agents"},
		{"assignment", []string{"task-a", "assignment"}, taskBase + "/assignment"},
		{"tool approvals", []string{"task-a", "tool-approvals"}, taskBase + "/tool-approvals"},
		{"tool decision", []string{"task-a", "tool-approvals", "apr-a", "decision"}, taskBase + "/tool-approvals/{approval_id}/decision"},
		{"threads", []string{"task-a", "threads"}, taskBase + "/threads"},
		{"thread messages", []string{"task-a", "threads", "thread-a", "messages"}, taskBase + "/threads/{thread_id}/messages"},
		{"assignments", []string{"task-a", "agent-assignments"}, taskBase + "/agent-assignments"},
		{"assignment agent", []string{"task-a", "agent-assignments", "agent-a"}, taskBase + "/agent-assignments/{agent_id}"},
		{"assignment stop", []string{"task-a", "agent-assignments", "agent-a", "stop"}, taskBase + "/agent-assignments/{agent_id}/stop"},
		{"stream subscription", []string{"task-a", "thread-stream-subscription"}, taskBase + "/thread-stream-subscription"},
		{"comments", []string{"task-a", "comments"}, taskBase + "/comments"},
		{"stop", []string{"task-a", "stop"}, taskBase + "/stop"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aiAgentClientTaskHTTPRoute(base, tt.segments); got != tt.want {
				t.Fatalf("route = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAIAgentClientTaskHTTPRouteRejectsUnknownTaskScopedRoutes(t *testing.T) {
	base := "/v2/client/workspaces/{workspace_id}/ai-agent"
	tests := []struct {
		name     string
		segments []string
	}{
		{"empty", nil},
		{"single task", []string{"task-a"}},
		{"wrong assigned profiles length", []string{"assigned-agent-profiles", "extra"}},
		{"unknown leaf", []string{"task-a", "unknown"}},
		{"wrong tool decision action", []string{"task-a", "tool-approvals", "apr-a", "approve"}},
		{"assignment wrong action", []string{"task-a", "agent-assignments", "agent-a", "pause"}},
		{"leaf too deep", []string{"task-a", "stop", "extra"}},
		{"thread wrong action", []string{"task-a", "threads", "thread-a", "reply"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aiAgentClientTaskHTTPRoute(base, tt.segments); got != "" {
				t.Fatalf("route = %q, want empty", got)
			}
		})
	}
}
