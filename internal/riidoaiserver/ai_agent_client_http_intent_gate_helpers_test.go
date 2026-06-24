package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func postIntentGateFollowup(t *testing.T, handler http.Handler, assigned AIAgentTaskActionResponse) AIAgentTaskActionResponse {
	t.Helper()
	body := `{"body":"우리 신기능 셀링 포인트 세 가지 반영해서 훅이 강한 카피라이팅 초안 3개만 짜줘."}`
	req := httptest.NewRequest(http.MethodPost,
		"/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/"+assigned.TaskID+"/threads/"+assigned.ThreadID+"/messages",
		strings.NewReader(body),
	)
	req.Header.Set(aiAgentTokenHeader, "user-token")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("followup status=%d body=%s", resp.Code, resp.Body.String())
	}
	var out AIAgentTaskActionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("followup json: %v", err)
	}
	return out
}
