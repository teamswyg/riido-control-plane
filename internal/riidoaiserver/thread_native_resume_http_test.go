package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPThreadFollowupReusesAuthoritativeNativeSessionAcrossTerminalStates(t *testing.T) {
	for _, terminal := range []AssignmentState{AssignmentCompleted, AssignmentFailed, AssignmentCancelled} {
		t.Run(string(terminal), func(t *testing.T) {
			f := newNativeResumeHTTPFixture(t)
			first := f.finishFirst(t, terminal)
			var next AIAgentTaskActionResponse
			f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "Continue this task in the same repository."}, http.StatusAccepted, &next)
			poll, err := f.store.PollAgent(context.Background(), "agent-owned-codex", PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
			if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID || poll.Assignment.ResumeSessionID != "native-http-session" {
				t.Fatal("real HTTP followup lost the exact selected thread native session")
			}
		})
	}
}

func TestHTTPSharedAgentFollowupKeepsNativeSessionWhenRequesterChanges(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	f.token = "native-teammate-token"
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "Continue the authorized shared task."}, http.StatusAccepted, &next)
	projection, found, err := f.store.LoadAssignmentProjection(context.Background(), next.AssignmentID)
	if err != nil || !found || projection.Assignment.CreatedBy != "user-2" || projection.Assignment.ResumeSessionID != "native-http-session" {
		t.Fatal("requester identity was mistaken for the immutable Agent owner")
	}
}

func (f *nativeResumeHTTPFixture) finishFirst(t *testing.T, terminal AssignmentState) AIAgentTaskActionResponse {
	t.Helper()
	var first AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", RequestID: "first-native-run"}, http.StatusAccepted, &first)
	poll, err := f.store.PollAgent(context.Background(), "agent-owned-codex", PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != first.AssignmentID {
		t.Fatal("first native run was not claimed")
	}
	terminalType := EventAssignmentCompleted
	switch terminal {
	case AssignmentFailed:
		terminalType = EventAssignmentFailed
	case AssignmentCancelled:
		terminalType = EventAssignmentCancelled
	}
	for _, phase := range []struct {
		state     AssignmentState
		kind, pin string
	}{{AssignmentReady, EventAssignmentReady, ""}, {AssignmentRunning, EventProviderSessionPinned, "native-http-session"}, {terminal, terminalType, "native-http-session"}} {
		f.post(t, "/v1/agents/agent-owned-codex/events", AgentEventRequest{AssignmentID: first.AssignmentID, TaskID: "task-new", DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: phase.state, EventType: phase.kind, ProviderSessionID: phase.pin}, http.StatusOK, nil)
	}
	return first
}
