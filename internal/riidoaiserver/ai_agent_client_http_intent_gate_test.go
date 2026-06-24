package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentIntentGateWaitsForUserBeforeDurableAssignment(t *testing.T) {
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
	assertIntentGateAction(t, assigned)
	pollReq := PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}
	pollNone, err := assignmentStore.PollAgent(t.Context(), assigned.AgentID, pollReq)
	if err != nil {
		t.Fatalf("PollAgent before followup: %v", err)
	}
	if pollNone.Action != PollNone || pollNone.Assignment != nil {
		t.Fatalf("intent-gated assignment must not reach daemon before user reply: %+v", pollNone)
	}
	followup := postIntentGateFollowup(t, handler, assigned)
	pollStart, err := assignmentStore.PollAgent(t.Context(), followup.AgentID, pollReq)
	if err != nil {
		t.Fatalf("PollAgent after followup: %v", err)
	}
	if pollStart.Action != PollStart || pollStart.Assignment == nil {
		t.Fatalf("followup assignment was not durable: %+v", pollStart)
	}
	if !strings.Contains(pollStart.Assignment.Prompt, "### New User Instruction") ||
		!strings.Contains(pollStart.Assignment.Prompt, "훅이 강한 카피라이팅 초안 3개") ||
		!strings.Contains(pollStart.Assignment.Prompt, "신기능 셀링 포인트 세 가지") {
		t.Fatalf("followup prompt missing latest instruction/document:\n%s", pollStart.Assignment.Prompt)
	}
}

func assertIntentGateAction(t *testing.T, assigned AIAgentTaskActionResponse) {
	t.Helper()
	if assigned.WorkStatus != AgentWorkStatusWaitingForUser ||
		assigned.AssignmentState != AgentAssignmentStateRunning ||
		assigned.CommentKind != AgentTaskCommentRuntimeProgress ||
		assigned.Message != clientMessageNeedUserInput ||
		!isIntentGateAssignmentID(assigned.AssignmentID) {
		t.Fatalf("intent gate action = %+v", assigned)
	}
}
