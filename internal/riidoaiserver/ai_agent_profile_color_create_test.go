package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientCreatePreservesRequestedProfileColor(t *testing.T) {
	for _, path := range []string{
		"/v1/client/ai-agent/agents",
		"/v2/client/workspaces/workspace-dev/ai-agent/agents",
	} {
		for _, color := range agentProfileColorOrder {
			t.Run(path+"/"+string(color), func(t *testing.T) {
				server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
					PrincipalID: "user-1", Token: "owner-token", Scopes: []string{"ai-agent:*"},
				}})
				body := fmt.Sprintf(`{"name":"preview color","runtime_id":"runtime-cursor-dev","visibility":"private","profile_color":%q}`, color)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
				req.Header.Set(aiAgentTokenHeader, "owner-token")
				response := httptest.NewRecorder()
				server.ServeHTTP(response, req)
				if response.Code != http.StatusCreated {
					t.Fatalf("create status=%d body=%s", response.Code, response.Body.String())
				}
				var created AgentClientRecordResponse
				if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
					t.Fatal(err)
				}
				if created.Agent.ProfileColor != color {
					t.Fatalf("profile_color=%q, want preview color %q", created.Agent.ProfileColor, color)
				}
			})
		}
	}
}

func TestHTTPAIAgentClientCreateRejectsUnknownProfileColor(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1", Token: "owner-token", Scopes: []string{"ai-agent:*"},
	}})
	req := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-dev/ai-agent/agents", strings.NewReader(
		`{"name":"invalid color","runtime_id":"runtime-cursor-dev","visibility":"private","profile_color":"red"}`,
	))
	req.Header.Set(aiAgentTokenHeader, "owner-token")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, req)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "profile_color must be") {
		t.Fatalf("invalid profile color status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestPersistentAIAgentClientCreateRetainsRequestedProfileColorAfterRestart(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	var req CreateAgentConfigurationRequest
	if err := json.Unmarshal([]byte(`{"name":"persist preview color","runtime_id":"runtime-cursor-dev","visibility":"private","profile_color":"purple"}`), &req); err != nil {
		t.Fatal(err)
	}
	created, err := store.CreateAIAgent(ctx, principal, req)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatal(err)
	}
	bootstrap, err := reopened.BootstrapAIAgentClient(ctx, principal, ClientKindWeb)
	if err != nil {
		t.Fatal(err)
	}
	agent, ok := findAIAgent(bootstrap.Agents, created.Agent.AgentID)
	if !ok || agent.ProfileColor != AgentProfileColorPurple {
		t.Fatalf("reopened profile_color=%q, found=%v, want purple", agent.ProfileColor, ok)
	}
}
