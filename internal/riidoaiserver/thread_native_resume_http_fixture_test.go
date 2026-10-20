package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type nativeResumeHTTPFixture struct {
	handler http.Handler
	store   *Store
	client  *DevelopmentAIAgentClientStore
	base    string
	token   string
}

func newNativeResumeHTTPFixture(t *testing.T) *nativeResumeHTTPFixture {
	t.Helper()
	client := NewDevelopmentAIAgentClientStore()
	client.mu.Lock()
	agent := client.agents["agent-owned-codex"]
	agent.WorkStatus, agent.AssignedTaskCount, agent.Visibility = AgentWorkStatusIdle, 0, AgentVisibilityPublic
	client.agents[agent.AgentID] = agent
	client.mu.Unlock()
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
	t.Cleanup(store.Close)
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "native-owner-token", Scopes: []string{"riido:*"}}, {PrincipalID: "user-2", Token: "native-teammate-token", Scopes: []string{"riido:*"}}})
	if err != nil {
		t.Fatal(err)
	}
	handler := NewServer(ServerConfig{Assignment: store, AIAgentClient: client, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: auth}).Handler()
	return &nativeResumeHTTPFixture{handler: handler, store: store, client: client, base: "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent", token: "native-owner-token"}
}

func (f *nativeResumeHTTPFixture) post(t *testing.T, path string, body any, status int, destination any) {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(encoded)))
	req.Header.Set(aiAgentTokenHeader, f.token)
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("native followup HTTP status=%d want=%d", response.Code, status)
	}
	if destination != nil {
		if err := json.Unmarshal(response.Body.Bytes(), destination); err != nil {
			t.Fatal(err)
		}
	}
}
