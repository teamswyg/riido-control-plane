package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentThreadProgressBatchIngestsAssignmentAndClientEvent(t *testing.T) {
	ctx := context.Background()
	assignmentStore := NewStore()
	defer assignmentStore.Close()
	assignment, err := assignmentStore.AssignTask(ctx, "task-1", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-owned-codex",
		RuntimeProvider: "codex",
		Prompt:          "summarize team projects",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if _, err := assignmentStore.PollAgent(ctx, "agent-owned-codex", PollRequest{DaemonID: "daemon-1", RuntimeID: "runtime-1"}); err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "daemon-1",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}, {
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:stream"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: NewDevelopmentAIAgentClientStore(),
		Authorizer:    authorizer,
	}).Handler()

	body := `{"assignment_id":"` + assignment.ID + `","task_id":"task-1","thread_id":"thread-task-1-codex-live","daemon_id":"daemon-1","runtime_id":"runtime-1","run_id":"run-1","lines":[{"seq":1,"message":"생각 중..."},{"seq":2,"message":"팀 프로젝트 수집 중 - 팀의 프로젝트 목록을 조회 중."}]}`
	ingestReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-owned-codex/thread-progress", strings.NewReader(body))
	ingestReq.Header.Set("Authorization", "Bearer daemon-token")
	ingestResp := httptest.NewRecorder()
	handler.ServeHTTP(ingestResp, ingestReq)
	if ingestResp.Code != http.StatusAccepted {
		t.Fatalf("ingest status=%d body=%s", ingestResp.Code, ingestResp.Body.String())
	}
	var response AgentThreadProgressBatchResponse
	if err := json.Unmarshal(ingestResp.Body.Bytes(), &response); err != nil {
		t.Fatalf("ingest json: %v", err)
	}
	if response.AcceptedLines != 2 || response.Event.EventType != AgentClientEventThreadProgress || response.Event.ThreadID != "thread-task-1-codex-live" || len(response.Event.Lines) != 2 {
		t.Fatalf("ingest response = %+v", response)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	handler.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK {
		t.Fatalf("events status=%d body=%s", eventsResp.Code, eventsResp.Body.String())
	}
	message := eventsResp.Body.String()
	if !strings.Contains(message, "event: agent_thread_progress\n") || !strings.Contains(message, "팀 프로젝트 수집 중") {
		t.Fatalf("events body = %q", message)
	}
}
