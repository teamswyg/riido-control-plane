package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestContextReadStreamRechecksResourceAccessDuringDelivery(t *testing.T) {
	for _, mode := range []string{"replay", "live"} {
		t.Run(mode, func(t *testing.T) {
			s, reader, auth, event := contextReadStreamServer(t)
			reader.value.Resources = []AIAgentTaskContextResource{{ID: "hidden-resource"}}
			checks := 0
			s.config.Authorizer = authorizationFunc(func(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
				if req.TaskID != "" {
					checks++
					if checks > 1 {
						// Fallback values can remain even when their section fails.
						reader.value.ResourceStatus = "unavailable"
					}
				}
				return auth.Authorize(ctx, token, req)
			})
			second := event
			second.Seq = 2
			body := contextReadStreamRequest(s, mode, event, second).Body.String()
			if strings.Count(body, "event: agent_context_read") != 2 || strings.Count(body, "hidden-resource") != 1 || strings.Contains(body, "omitted_sources") {
				t.Fatalf("revoked resource leaked: %s", body)
			}
		})
	}
}

func TestContextReadStreamUsesCurrentRolesAndIdentity(t *testing.T) {
	for _, mode := range []string{"replay", "live"} {
		for _, change := range []string{"role", "identity", "workspace"} {
			t.Run(mode+"/"+change, func(t *testing.T) {
				s, _, auth, event := contextReadStreamServer(t)
				store := s.aiAgent.(*DevelopmentAIAgentClientStore)
				agent := store.agents["agent-owned-codex"]
				agent.Visibility = AgentVisibilityPrivate
				store.agents[agent.AgentID] = agent
				s.config.Authorizer = authorizationFunc(func(ctx context.Context, token string, req AuthorizationRequest) (AuthorizationResult, error) {
					current, err := auth.Authorize(ctx, token, req)
					current.Roles = []AgentCatalogRole{AgentCatalogRoleAdmin}
					if req.TaskID != "" {
						switch change {
						case "role":
							current.Roles = nil
						case "identity":
							current.PrincipalID = "someone-else"
						case "workspace":
							current.WorkspaceID = "elsewhere"
						}
					}
					return current, err
				})
				if body := contextReadStreamRequest(s, mode, event).Body.String(); body != "" {
					t.Fatalf("stale subscription privilege leaked: %q", body)
				}
			})
		}
	}
}
