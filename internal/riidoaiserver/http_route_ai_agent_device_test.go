package riidoaiserver

import "testing"

func TestAIAgentClientDeviceHTTPRouteClassifiesDaemonRoutes(t *testing.T) {
	base := "/v2/client/workspaces/{workspace_id}/ai-agent"
	tests := []struct {
		name     string
		segments []string
		want     string
	}{
		{"list devices", []string{"devices"}, base + "/devices"},
		{"device daemon", []string{"devices", "device-a", "daemon"}, base + "/devices/{device_id}/daemon"},
		{"device daemons", []string{"devices", "device-a", "daemons"}, base + "/devices/{device_id}/daemons"},
		{"device start", []string{"devices", "device-a", "daemon", "start"}, base + "/devices/{device_id}/daemon/{action}"},
		{"device restart", []string{"devices", "device-a", "daemon", "restart"}, base + "/devices/{device_id}/daemon/{action}"},
		{"device stop", []string{"devices", "device-a", "daemon", "stop"}, base + "/devices/{device_id}/daemon/{action}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aiAgentClientDeviceHTTPRoute(base, tt.segments); got != tt.want {
				t.Fatalf("route = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAIAgentClientAgentHTTPRouteClassifiesDaemonRoutes(t *testing.T) {
	base := "/v2/client/workspaces/{workspace_id}/ai-agent"
	tests := []struct {
		name     string
		segments []string
		want     string
	}{
		{"list agents", []string{"agents"}, base + "/agents"},
		{"agent detail", []string{"agents", "agent-a"}, base + "/agents/{agent_id}"},
		{"agent daemon", []string{"agents", "agent-a", "daemon"}, base + "/agents/{agent_id}/daemon"},
		{"agent restart", []string{"agents", "agent-a", "daemon", "restart"}, base + "/agents/{agent_id}/daemon/{action}"},
		{"agent editability", []string{"agents", "agent-a", "editability"}, base + "/agents/{agent_id}/editability"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aiAgentClientAgentHTTPRoute(base, tt.segments); got != tt.want {
				t.Fatalf("route = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAIAgentClientDaemonHTTPRouteRejectsUnknownActions(t *testing.T) {
	base := "/v2/client/workspaces/{workspace_id}/ai-agent"
	tests := [][]string{
		{"devices", "device-a", "daemon", "pause"},
		{"devices", "device-a", "daemons", "restart"},
		{"agents", "agent-a", "daemon", "pause"},
		{"agents", "agent-a", "editability", "restart"},
	}
	for _, segments := range tests {
		if got := aiAgentClientRouteFromSegments(base, segments); got != "" {
			t.Fatalf("route(%v) = %q, want empty", segments, got)
		}
	}
}
