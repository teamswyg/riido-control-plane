package riidoaiserver

import "strings"

const riidoClientSurfaceTraceKey = "riido.client.surface"

func traceHTTPClientSurface(route, path, userAgent string) string {
	route = strings.TrimSpace(route)
	path = strings.TrimSpace(path)
	userAgent = strings.ToLower(strings.TrimSpace(userAgent))
	switch {
	case strings.HasPrefix(route, "/v1/daemon/"):
		return "daemon"
	case strings.HasPrefix(route, "/v1/agents/"):
		return "daemon"
	case strings.Contains(userAgent, "riido-desktop"):
		return "desktop"
	case strings.Contains(route, "/ai-agent/devices"):
		return "desktop_candidate"
	case strings.HasPrefix(route, "/v1/client/"):
		return "client_app"
	case strings.HasPrefix(route, "/v2/client/"):
		return "client_app"
	case strings.HasPrefix(route, "/v3/client/"):
		return "client_app"
	case strings.HasPrefix(route, "/v1/component-tasks/"):
		return "component_integration"
	case path == "/healthz" || path == "/readyz":
		return "healthcheck"
	default:
		return "unknown"
	}
}
