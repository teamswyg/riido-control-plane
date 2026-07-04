package riidoaiserver

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"testing"
)

const (
	riidoAIServerRouteGoldenHash  = "009ba0eafb48921d94584788ddb1aea4206f7621dd87666bd4408bd0dc377a1f"
	riidoAIServerStatusGoldenHash = "7191fa30926ca05804763741bfb752231ea4ead921f9a58640b875001038437f"
)

func TestRiidoAIServerBehaviorGolden(t *testing.T) {
	routes := riidoAIServerRouteGoldenPayload(t)
	if got := sha256String(routes); got != riidoAIServerRouteGoldenHash {
		t.Fatalf("route surface hash = %s, want %s\n%s", got, riidoAIServerRouteGoldenHash, routes)
	}
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*"},
	}})
	statuses := riidoAIServerStatusGoldenPayload(t, server)
	if got := sha256String(statuses); got != riidoAIServerStatusGoldenHash {
		t.Fatalf("status surface hash = %s, want %s\n%s", got, riidoAIServerStatusGoldenHash, statuses)
	}
}

func riidoAIServerRouteGoldenPayload(t *testing.T) string {
	t.Helper()
	patterns := make([]string, 0, len(serverRoutes))
	for _, route := range serverRoutes {
		if route.pattern == "" || route.handler == nil {
			t.Fatalf("incomplete route binding: %+v", route)
		}
		patterns = append(patterns, route.pattern)
	}
	sort.Strings(patterns)
	if len(patterns) != 24 {
		t.Fatalf("route count = %d, want 24\n%s", len(patterns), strings.Join(patterns, "\n"))
	}
	return strings.Join(patterns, "\n") + "\n"
}

func sha256String(value string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(value))) }
