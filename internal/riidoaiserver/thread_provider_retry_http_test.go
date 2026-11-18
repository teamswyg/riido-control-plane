package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestHTTPCodexReplyAfterConfirmedPreStartFailure(t *testing.T) {
	for _, repo := range []bool{true, false} {
		name := "no repository"
		if repo {
			name = "with repository"
		}
		t.Run(name, func(t *testing.T) { testCodexPreStartReply(t, repo) })
	}
}

func testCodexPreStartReply(t *testing.T, repo bool) {
	t.Helper()
	f := newNativeResumeHTTPFixture(t)
	if !repo {
		f.taskContext.contextSnapshot.Repositories = nil
	}
	var first AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", RequestID: "first", Instruction: "Review the linked PR."}, http.StatusAccepted, &first)
	poll := PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}
	leased, err := f.store.PollAgent(t.Context(), first.AgentID, poll)
	if err != nil || leased.Assignment == nil {
		t.Fatal("missing lease", err)
	}
	event := AgentEventRequest{AssignmentID: first.AssignmentID, TaskID: first.TaskID, DaemonID: poll.DaemonID, DeviceID: poll.DeviceID, RuntimeID: poll.RuntimeID}
	events := "/v1/agents/" + first.AgentID + "/events"
	event.State, event.EventType = AssignmentReady, EventAssignmentReady
	f.post(t, events, event, http.StatusOK, nil)
	event.EventType = EventRiidoLog
	event.Metadata = (assignment.RecoveryContext{State: "unavailable", Reason: "budget_exhausted", Provider: "codex", OutcomeCertainty: "not_started", Attempt: 3}).Metadata()
	event.Metadata["lease_token"] = leased.Assignment.LeaseToken
	event.Metadata[metadatakeys.AssignmentEventKey.String()] = "preflight-unavailable"
	f.post(t, events, event, http.StatusOK, nil)
	event.State, event.EventType, event.Metadata = AssignmentFailed, EventAssignmentFailed, nil
	f.post(t, events, event, http.StatusOK, nil)
	input := CreateAIAgentTaskThreadMessageRequest{Body: "PR #1193", RequestID: "reply", SourceMessageID: "reply-comment"}
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	f.token = "native-teammate-token"
	f.post(t, path, input, http.StatusForbidden, nil)
	f.token = "native-owner-token"
	var next, replay AIAgentTaskActionResponse
	f.post(t, path, input, http.StatusAccepted, &next)
	f.post(t, path, input, http.StatusAccepted, &replay)
	f.post(t, path, CreateAIAgentTaskThreadMessageRequest{Body: "Another instruction", RequestID: "other"}, http.StatusConflict, nil)
	if next.AssignmentID == first.AssignmentID || next.AssignmentID != replay.AssignmentID {
		t.Fatal("reply did not create exactly one new execution")
	}
	work, err := f.store.PollAgent(t.Context(), first.AgentID, poll)
	if err != nil || work.Assignment == nil || work.Assignment.ID != next.AssignmentID || work.Assignment.ResumeSessionID != "" || !strings.Contains(work.Assignment.Prompt, input.Body) {
		t.Fatal("new execution missing reply", err)
	}
	history, err := f.client.ListAIAgentTaskThreadHistory(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, first.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range history.Threads {
		if row.ThreadID == next.ThreadID && row.ParentThreadID == first.ThreadID && row.ConversationID == first.ThreadID {
			return
		}
	}
	t.Fatal("reply lost original conversation")
}
