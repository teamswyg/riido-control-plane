package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentAssignmentExecutesWithoutIntentGate(t *testing.T) {
	handler, assignmentStore := newIntentGateHTTPTestServer(t)
	assignReq := httptest.NewRequest(http.MethodPost,
		"/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-copy/agent-assignments",
		strings.NewReader(`{"agent_id":"agent-owned-codex"}`),
	)
	assignReq.Header.Set(aiAgentTokenHeader, "user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	assertDirectAssignmentAction(t, assigned)
	pollStart, err := assignmentStore.PollAgent(t.Context(), assigned.AgentID, intentDialoguePollRequest())
	if err != nil || pollStart.Action != PollStart || pollStart.Assignment == nil {
		t.Fatalf("initial assignment must reach daemon: %+v %v", pollStart, err)
	}
	if !strings.Contains(pollStart.Assignment.Prompt, marketingTaskContextFixture().Document.Content) {
		t.Fatal("initial task document missing")
	}
}

func assertDirectAssignmentAction(t *testing.T, assigned AIAgentTaskActionResponse) {
	t.Helper()
	if assigned.WorkStatus == AgentWorkStatusWaitingForUser || assigned.CommentKind == AgentTaskCommentNeedsInput || assigned.ActiveStream == nil || isIntentGateAssignmentID(assigned.AssignmentID) {
		t.Fatalf("assignment must execute without synthetic intent question: %+v", assigned)
	}
}
