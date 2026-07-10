package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentIntentGateExecutesExplicitMetadataTitle(t *testing.T) {
	t.Parallel()
	handler, assignmentStore := newExplicitMetadataHTTPTestServer(t)
	assignReq := httptest.NewRequest(http.MethodPost,
		"/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-fill-ai-content/agent-assignments",
		strings.NewReader(`{"agent_id":"agent-owned-codex"}`),
	)
	assignReq.Header.Set(aiAgentTokenHeader, "user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	assigned := decodeAIAgentTaskActionResponse(t, assignResp.Body.Bytes())
	if assigned.WorkStatus != AgentWorkStatusIdle ||
		assigned.AssignmentState != AgentAssignmentStateQueued ||
		assigned.CommentKind != "" ||
		isIntentGateAssignmentID(assigned.AssignmentID) ||
		assigned.ActiveStream == nil {
		t.Fatalf("explicit metadata title should create durable assignment: %+v", assigned)
	}
	pollStart, err := assignmentStore.PollAgent(t.Context(), assigned.AgentID, PollRequest{
		DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if pollStart.Action != PollStart || pollStart.Assignment == nil {
		t.Fatalf("explicit metadata assignment was not durable: %+v", pollStart)
	}
	if !strings.Contains(pollStart.Assignment.Prompt, "- intent_class: explicit_instruction") ||
		!strings.Contains(pollStart.Assignment.Prompt, "- title: AI 내용 채우기") {
		t.Fatalf("poll prompt missing explicit metadata evidence:\n%s", pollStart.Assignment.Prompt)
	}
}

func newExplicitMetadataHTTPTestServer(t *testing.T) (http.Handler, *Store) {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	return NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext: &assignmentHTTPRequestTaskContextReader{contextSnapshot: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID: "task-fill-ai-content", ComponentType: "task", Title: "AI 내용 채우기",
			},
		}},
		Authorizer: authorizer,
	}).Handler(), assignmentStore
}
