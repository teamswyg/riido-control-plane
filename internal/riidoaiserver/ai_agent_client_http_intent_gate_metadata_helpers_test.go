package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func newMetadataIntentGateHTTPTestServer(t *testing.T) (http.Handler, *Store) {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   &assignmentHTTPRequestTaskContextReader{contextSnapshot: metadataIntentTaskContextFixture()},
		Authorizer:    authorizer,
	}).Handler()
	return server, assignmentStore
}

func metadataIntentTaskContextFixture() AIAgentTaskContext {
	return AIAgentTaskContext{Component: AIAgentTaskContextComponent{
		ID:            "task-metadata",
		ComponentType: "task",
		Title:         "[1.23 신기능 마케팅] 카피라이트 3개안 준비",
	}}
}

func decodeAIAgentTaskActionResponse(t *testing.T, body []byte) AIAgentTaskActionResponse {
	t.Helper()
	var out AIAgentTaskActionResponse
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode action response: %v body=%s", err, string(body))
	}
	return out
}
