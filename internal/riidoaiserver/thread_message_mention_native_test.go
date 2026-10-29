package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestHTTPThreadReplyHandoffKeepsEachAgentsNativeSession(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	other := f.client.agents[first.AgentID]
	other.AgentID, other.Name = "agent-second-codex", "Second Codex"
	other.WorkStatus, other.AssignedTaskCount = AgentWorkStatusIdle, 0
	f.client.agents[other.AgentID] = other
	var second AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: mentionHTML(other.AgentID) + " continue", RequestID: "handoff"}, http.StatusAccepted, &second)
	poll, err := f.store.PollAgent(t.Context(), other.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != second.AssignmentID || poll.Assignment.ResumeSessionID != "" {
		t.Fatal("same-provider handoff claimed the original Agent's native session")
	}
	for _, phase := range []struct {
		state AssignmentState
		kind  string
	}{{AssignmentReady, EventAssignmentReady}, {AssignmentRunning, EventProviderSessionPinned}, {AssignmentCompleted, EventAssignmentCompleted}} {
		f.post(t, "/v1/agents/"+other.AgentID+"/events", AgentEventRequest{AssignmentID: second.AssignmentID, TaskID: first.TaskID, DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: phase.state, EventType: phase.kind, ProviderSessionID: "second-native-session", Message: "Second Agent reviewed the blog template"}, http.StatusOK, nil)
	}
	for _, tc := range []struct{ target, session string }{{first.AgentID, "native-http-session"}, {other.AgentID, "second-native-session"}} {
		var next AIAgentTaskActionResponse
		f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: mentionHTML(tc.target) + " review the result", RequestID: tc.target}, http.StatusAccepted, &next)
		projection, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
		if err != nil || !found || projection.Assignment.AgentID != tc.target || projection.Assignment.ResumeSessionID != tc.session {
			t.Fatal("returning participant resumed the wrong session")
		}
		if !strings.Contains(projection.Assignment.Prompt, "Second Agent reviewed the blog template") {
			t.Fatal("continuation lost the other participant's linked conversation context")
		}
	}
}
