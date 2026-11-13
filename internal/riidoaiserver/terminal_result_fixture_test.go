package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

const terminalResultBody = "검토를 완료했습니다. 기존 API URL은 유지되며 새 DB 연결을 검증했습니다."

// Commit the terminal journal event without delivering its client fanout.
func terminalResultFixture(t *testing.T) (*nativeResumeHTTPFixture, AIAgentTaskActionResponse, AgentEventRequest, TaskEvent) {
	t.Helper()
	f := newNativeResumeHTTPFixture(t)
	var first AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", RequestID: "terminal-result"}, http.StatusAccepted, &first)
	if _, err := f.store.PollAgent(t.Context(), first.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}); err != nil {
		t.Fatal(err)
	}
	req := AgentEventRequest{AssignmentID: first.AssignmentID, TaskID: first.TaskID, DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: AssignmentRunning, EventType: EventAssignmentRunning}
	f.post(t, "/v1/agents/"+first.AgentID+"/events", req, http.StatusOK, nil)
	req.State, req.EventType, req.Message = AssignmentCompleted, EventAssignmentCompleted, terminalResultBody
	result, err := f.store.RecordAgentEvent(t.Context(), first.AgentID, req)
	if err != nil {
		t.Fatal(err)
	}
	return f, first, req, result.Event
}

func readTerminalResultHistory(t *testing.T, f *nativeResumeHTTPFixture) AIAgentTaskThreadHistoryCollectionResponse {
	t.Helper()
	body := aiAgentSmokeRequest(t, f.handler, http.MethodGet, strings.Replace(f.base, "/v2/", "/v3/", 1)+"/tasks/task-new/threads", f.token, "", http.StatusOK)
	var history AIAgentTaskThreadHistoryCollectionResponse
	aiAgentSmokeDecode(t, body, &history)
	return history
}

func assertTerminalResultHistory(t *testing.T, history AIAgentTaskThreadHistoryCollectionResponse) {
	t.Helper()
	if len(history.Threads) != 1 || history.ActiveStream != nil || history.Threads[0].AssignmentState != AgentAssignmentStateCompleted {
		t.Fatalf("completed thread was not preserved: %+v", history)
	}
	count := 0
	for _, message := range history.Threads[0].Messages {
		if message.Role == AIAgentTaskThreadMessageRoleAgent && message.Body == terminalResultBody {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("want original terminal result exactly once, got %d: %+v", count, history.Threads[0].Messages)
	}
}
