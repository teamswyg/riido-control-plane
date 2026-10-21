package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDeniedDeleteDoesNotCancelAssignments(t *testing.T) {
	for _, visibility := range []AgentVisibility{AgentVisibilityPublic, AgentVisibilityPrivate} {
		t.Run(string(visibility), func(t *testing.T) {
			store := NewDevelopmentAIAgentClientStore()
			agent := store.agents["agent-owned-codex"]
			agent.Visibility = visibility
			store.agents[agent.AgentID] = agent
			authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{
				{PrincipalID: "user-1", Token: "owner-token", Scopes: []string{"ai-agent:*", "task:task-delete-authorization:assign"}},
				{PrincipalID: "user-2", Token: "other-token", Scopes: []string{"ai-agent:*"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: store})
			t.Cleanup(assignments.Close)
			server := NewServer(ServerConfig{
				AIAgentClient: store, Assignment: assignments, Authorizer: authorizer,
				TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
			}).Handler()
			request := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-delete-authorization/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
			request.Header.Set("Authorization", "Bearer owner-token")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != http.StatusAccepted {
				t.Fatalf("assign status=%d body=%s", response.Code, response.Body.String())
			}
			var assigned AIAgentTaskActionResponse
			if err := json.Unmarshal(response.Body.Bytes(), &assigned); err != nil {
				t.Fatal(err)
			}
			poll := PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}
			if result, err := assignments.PollAgent(t.Context(), assigned.AgentID, poll); err != nil || result.Action != PollStart {
				t.Fatalf("start=%+v err=%v", result, err)
			}
			before, _, err := assignments.LoadAssignmentProjection(t.Context(), assigned.AssignmentID)
			if err != nil {
				t.Fatal(err)
			}
			request = httptest.NewRequest(http.MethodDelete, "/v1/client/ai-agent/agents/"+assigned.AgentID, nil)
			request.Header.Set("Authorization", "Bearer other-token")
			response = httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != http.StatusNotFound {
				t.Fatalf("delete status=%d body=%s", response.Code, response.Body.String())
			}
			after, found, err := assignments.LoadAssignmentProjection(t.Context(), assigned.AssignmentID)
			if err != nil || !found || after.Assignment.State != before.Assignment.State {
				t.Fatalf("denied delete changed assignment: before=%s after=%s found=%v err=%v", before.Assignment.State, after.Assignment.State, found, err)
			}
			if _, exists := store.agents[assigned.AgentID]; !exists {
				t.Fatal("denied delete removed agent")
			}
		})
	}
}
