package riidoaiserver

import (
	"net/http"
	"testing"
)

func newIntentGateHTTPTestServer(t *testing.T) (http.Handler, *Store) {
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
		TaskContext:   &assignmentHTTPRequestTaskContextReader{contextSnapshot: marketingTaskContextFixture()},
		Authorizer:    authorizer,
	}).Handler()
	return server, assignmentStore
}

func marketingTaskContextFixture() AIAgentTaskContext {
	return AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{
			ID:            "task-copy",
			ComponentType: "task",
			Title:         "[1.23 신기능 마케팅] 카피라이트 3개안 준비",
		},
		Document: AIAgentTaskContextDocument{
			Content:       "신기능 셀링 포인트 세 가지를 분석하고 마케팅 방향을 정리한다.",
			ContentFormat: "html",
		},
	}
}
