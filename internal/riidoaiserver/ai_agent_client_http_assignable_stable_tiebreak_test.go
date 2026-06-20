package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentAssignableAgentsUseStableIDTieBreak(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	ownedCodex := store.agents["agent-owned-codex"]
	ownedClaude := store.agents["agent-owned-claude"]
	ownedCodex.Name = "중복 이름"
	ownedClaude.Name = "중복 이름"
	store.agents[ownedCodex.AgentID] = ownedCodex
	store.agents[ownedClaude.AgentID] = ownedClaude

	publicA := store.agents["agent-public-openclaw"]
	publicA.AgentID = "agent-public-alpha"
	publicA.Name = "공개 중복"
	publicZ := publicA
	publicZ.AgentID = "agent-public-zeta"
	store.agents[publicA.AgentID] = publicA
	store.agents[publicZ.AgentID] = publicZ

	handler := NewServer(ServerConfig{
		AIAgentClient: store,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*", "task:task-1:read"}, "user-1"),
	}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-1/assignable-agents", nil)
	req.Header.Set("Authorization", "Bearer ai-agent-token")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("assignable status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AgentClientListResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("assignable json: %v", err)
	}
	if got, want := aiAgentIDsWithName(out.Agents, "중복 이름"), []string{"agent-owned-claude", "agent-owned-codex"}; !sameStrings(got, want) {
		t.Fatalf("owned duplicate-name order = %v, want %v; all agents=%+v", got, want, out.Agents)
	}
	if got, want := aiAgentIDsWithName(out.Agents, "공개 중복"), []string{"agent-public-alpha", "agent-public-zeta"}; !sameStrings(got, want) {
		t.Fatalf("public duplicate-name order = %v, want %v; all agents=%+v", got, want, out.Agents)
	}
}
