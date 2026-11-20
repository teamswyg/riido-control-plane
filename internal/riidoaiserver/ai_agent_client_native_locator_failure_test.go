package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientNativeLocatorFailureDoesNotPublishAssignment(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*", "task:task-smoke:assign", "agent:agent-owned-codex:poll"}}})
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.NotFoundHandler())
	defer upstream.Close()
	reader, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
	if err != nil {
		t.Fatal(err)
	}
	agents := NewDevelopmentAIAgentClientStore()
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: agents})
	defer assignments.Close()
	handler := NewServer(ServerConfig{AIAgentClient: agents, Assignment: assignments, TaskContext: reader, Authorizer: authorizer}).Handler()
	request := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-smoke/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex","instruction":"Inspect only"}`))
	request.Header.Set(aiAgentTokenHeader, "user-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code < 400 || response.Code >= 600 {
		t.Fatalf("native lookup failure accepted: status=%d body=%s", response.Code, response.Body.String())
	}
	poll := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`))
	poll.Header.Set("Authorization", "Bearer user-token")
	polled := httptest.NewRecorder()
	handler.ServeHTTP(polled, poll)
	var result PollResponse
	if polled.Code != http.StatusOK || json.Unmarshal(polled.Body.Bytes(), &result) != nil || result.Action == PollStart || result.Assignment != nil {
		t.Fatalf("failed native context left dispatchable work: status=%d body=%s", polled.Code, polled.Body.String())
	}
}
