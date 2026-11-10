package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
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
			if !strings.Contains(poll.Assignment.Prompt, "Continue this task in the same repository.") || strings.Contains(poll.Assignment.Prompt, "# Riido AI Agent Assignment") || strings.Contains(poll.Assignment.Prompt, "Previous Thread Message") {
				t.Fatalf("resumed prompt repeated the assignment instead of the new instruction: %s", poll.Assignment.Prompt)
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

func TestHTTPThreadFollowupKeepsWorkdirWhenPullRequestIsLinked(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	repositories := f.taskContext.contextSnapshot.Repositories
	f.taskContext.contextSnapshot.Repositories = nil
	first := f.finishFirst(t, AssignmentCompleted)

	f.taskContext.contextSnapshot.Repositories = repositories
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{
		Body:      "Continue in the same conversation after linking the pull request.",
		RequestID: "followup-after-pr-link",
	}, http.StatusAccepted, &next)
	projection, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
	if err != nil || !found || projection.Assignment.ResumeSessionID != "native-http-session" || projection.Assignment.Worktree != nil {
		t.Fatal("follow-up changed the original session's execution binding", err)
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

func TestHTTPThreadFollowupWithoutNativeSessionFailsExplicitly(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	var first AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", RequestID: "legacy-run"}, http.StatusAccepted, &first)
	if _, err := f.store.PollAgent(t.Context(), first.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []struct {
		state AssignmentState
		kind  string
	}{{AssignmentReady, EventAssignmentReady}, {AssignmentRunning, EventAssignmentRunning}, {AssignmentCompleted, EventAssignmentCompleted}} {
		f.post(t, "/v1/agents/"+first.AgentID+"/events", AgentEventRequest{AssignmentID: first.AssignmentID, TaskID: first.TaskID, DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev", State: phase.state, EventType: phase.kind}, http.StatusOK, nil)
	}
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "Continue legacy work"}, http.StatusConflict, nil)
	projection, found, err := f.store.LoadAssignmentProjection(t.Context(), first.AssignmentID)
	if err != nil || !found || !strings.Contains(projection.Assignment.Prompt, "# Riido AI Agent Assignment") {
		t.Fatal("initial request lost full bootstrap context", err)
	}
}
