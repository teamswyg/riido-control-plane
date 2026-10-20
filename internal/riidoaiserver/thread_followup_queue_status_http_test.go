package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSingleFollowupDoesNotCarryBusyMessageIntoRunningHistory(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages",
		CreateAIAgentTaskThreadMessageRequest{Body: "Continue this single task."}, http.StatusAccepted, &next)
	projection, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
	if err != nil || !found || projection.Assignment.BlockedByAssignmentID != "" {
		t.Fatalf("expected an unblocked follow-up: found=%v err=%v", found, err)
	}
	poll, err := f.store.PollAgent(t.Context(), "agent-owned-codex", PollRequest{
		DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev",
	})
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID {
		t.Fatalf("follow-up was not leased: %v", err)
	}
	for _, state := range []AssignmentState{AssignmentReady, AssignmentRunning} {
		eventType := EventAssignmentReady
		if state == AssignmentRunning {
			eventType = EventAssignmentRunning
		}
		f.post(t, "/v1/agents/agent-owned-codex/events", AgentEventRequest{
			AssignmentID: next.AssignmentID, TaskID: "task-new", DaemonID: "daemon-dev-macbook",
			DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: state, EventType: eventType,
		}, http.StatusOK, nil)
		req := httptest.NewRequest(http.MethodGet,
			strings.Replace(f.base, "/v2/", "/v3/", 1)+"/tasks/task-new/threads", nil)
		req.Header.Set(aiAgentTokenHeader, f.token)
		out := httptest.NewRecorder()
		f.handler.ServeHTTP(out, req)
		if out.Code != http.StatusOK || strings.Contains(out.Body.String(), clientMessageAgentBusyQueued) {
			t.Fatalf("%s history retained a false busy message (HTTP %d)", state, out.Code)
		}
		if !strings.Contains(out.Body.String(), clientMessageTaskRunning) {
			t.Fatalf("%s history omitted the running status", state)
		}
	}
}
