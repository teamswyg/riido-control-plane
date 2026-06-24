package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentIntentGateUsesMetadataWhenDocumentMissing(t *testing.T) {
	t.Parallel()
	handler, assignmentStore := newMetadataIntentGateHTTPTestServer(t)
	assignReq := httptest.NewRequest(http.MethodPost,
		"/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-metadata/agent-assignments",
		strings.NewReader(`{"agent_id":"agent-owned-codex"}`),
	)
	assignReq.Header.Set(aiAgentTokenHeader, "user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	assigned := decodeAIAgentTaskActionResponse(t, assignResp.Body.Bytes())
	assertIntentGateAction(t, assigned)
	pollReq := PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}
	pollNone, err := assignmentStore.PollAgent(t.Context(), assigned.AgentID, pollReq)
	if err != nil {
		t.Fatalf("PollAgent before followup: %v", err)
	}
	if pollNone.Action != PollNone || pollNone.Assignment != nil {
		t.Fatalf("metadata intent gate must not reach daemon before reply: %+v", pollNone)
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
		!strings.Contains(pollStart.Assignment.Prompt, "not provided") {
		t.Fatalf("metadata followup prompt missing intent evidence:\n%s", pollStart.Assignment.Prompt)
	}
}
