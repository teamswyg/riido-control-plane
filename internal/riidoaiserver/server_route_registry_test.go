package riidoaiserver

import "testing"

func TestServerRouteRegistryHasUniquePatterns(t *testing.T) {
	seen := map[string]bool{}
	for _, route := range serverRoutes {
		if route.pattern == "" || route.handler == nil {
			t.Fatalf("incomplete route binding: %+v", route)
		}
		if seen[route.pattern] {
			t.Fatalf("duplicate route pattern %q", route.pattern)
		}
		seen[route.pattern] = true
	}
}

func TestServerRouteRegistryIncludesCoreSurfaces(t *testing.T) {
	patterns := serverRoutePatterns()
	for _, want := range []string{
		"/v2/client/workspaces/",
		"/v3/client/workspaces/",
		"/v1/client/ai-agent/tasks/",
		"/v1/client/ai-agent/events",
		"/v1/daemon/runtime-snapshot",
		"/v1/agents/",
	} {
		if !patterns[want] {
			t.Fatalf("server route registry missing %s", want)
		}
	}
}

func serverRoutePatterns() map[string]bool {
	out := map[string]bool{}
	for _, route := range serverRoutes {
		out[route.pattern] = true
	}
	return out
}
