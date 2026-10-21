package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentAssignmentDeliversMetadataWhenDocumentMissing(t *testing.T) {
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
	assertDirectAssignmentAction(t, assigned)
	pollStart, err := assignmentStore.PollAgent(t.Context(), assigned.AgentID, intentDialoguePollRequest())
	if err != nil || pollStart.Action != PollStart || pollStart.Assignment == nil {
		t.Fatalf("metadata assignment must reach model: %+v %v", pollStart, err)
	}
	if !strings.Contains(pollStart.Assignment.Prompt, "not provided") {
		t.Fatal("missing document must be explicit in prompt")
	}
}
