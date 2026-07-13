package riidoaiserver

import (
	"net/http"
	"strings"
)

func traceHTTPRoute(method, path string) string {
	path = strings.TrimSpace(path)
	for _, route := range []func(string, string) string{
		agentHTTPRoute,
		daemonHTTPRoute,
		aiAgentClientHTTPRoute,
		componentTaskHTTPRoute,
	} {
		if pattern := route(method, path); pattern != "" {
			return pattern
		}
	}
	return ""
}

func agentHTTPRoute(method, path string) string {
	if method != http.MethodPost || !strings.HasPrefix(path, "/v1/agents/") {
		return ""
	}
	switch {
	case strings.HasSuffix(path, "/poll"):
		return "/v1/agents/{agent_id}/poll"
	case strings.HasSuffix(path, "/events"):
		return "/v1/agents/{agent_id}/events"
	case strings.HasSuffix(path, "/thread-progress"):
		return "/v1/agents/{agent_id}/thread-progress"
	default:
		return ""
	}
}

func daemonHTTPRoute(method, path string) string {
	switch {
	case method == http.MethodPost && path == "/v1/daemon/runtime-snapshot":
		return "/v1/daemon/runtime-snapshot"
	case method == http.MethodGet && path == "/v1/daemon/agent-bindings":
		return "/v1/daemon/agent-bindings"
	default:
		return ""
	}
}

func componentTaskHTTPRoute(method, path string) string {
	if !strings.HasPrefix(path, "/v1/component-tasks/") {
		return ""
	}
	_, suffix, ok := splitResourcePath(path, "/v1/component-tasks/")
	if method == http.MethodGet && ok && suffix == "events" {
		return "/v1/component-tasks/{task_id}/events"
	}
	if method == http.MethodPost {
		return "/v1/component-tasks/{task_id}"
	}
	return ""
}
