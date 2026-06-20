package riidoaiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientAssignFallsBackWhenTaskContextUnavailable(t *testing.T) {
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-smoke:assign", "agent:agent-owned-codex:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	taskContext := &assignmentHTTPRequestTaskContextReader{err: errors.New("private task context returned HTTP 401")}
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   taskContext,
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-smoke/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
	assignReq.Header.Set(aiAgentTokenHeader, "user-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	if len(taskContext.requests) != 1 {
		t.Fatalf("task context requests = %+v", taskContext.requests)
	}

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/poll", strings.NewReader(`{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`))
	pollReq.Header.Set("Authorization", "Bearer user-token")
	pollResp := httptest.NewRecorder()
	server.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusOK {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	var pollOut PollResponse
	if err := json.Unmarshal(pollResp.Body.Bytes(), &pollOut); err != nil {
		t.Fatalf("poll json: %v", err)
	}
	if pollOut.Action != PollStart || pollOut.Assignment == nil {
		t.Fatalf("poll response = %+v", pollOut)
	}
	if !strings.Contains(pollOut.Assignment.Prompt, "Task context was not available when this assignment was created.") ||
		!strings.Contains(pollOut.Assignment.Prompt, "- task_id: task-smoke") {
		t.Fatalf("assignment prompt = %s", pollOut.Assignment.Prompt)
	}
	if strings.Contains(pollOut.Assignment.Prompt, "private task context returned HTTP 401") {
		t.Fatalf("assignment prompt leaked task context error: %s", pollOut.Assignment.Prompt)
	}
}
