package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/assignment"
)

func TestHTTPWorkspaceRetryPreservesConversationAndInstruction(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	f.taskContext.contextSnapshot.Repositories[0].Source = "pull_request"
	f.taskContext.contextSnapshot.Component.KeyNumber = "RIID-6101"
	f.taskContext.contextSnapshot.Component.BranchName = "RIID-6101-missing-task-branch"
	repository := f.taskContext.contextSnapshot.Repositories[0]
	f.taskContext.contextSnapshot.linkedPullRequests = taskCheckoutPullRequests([]json.RawMessage{
		json.RawMessage(`{"number":1396,"htmlUrl":"https://github.com/` + repository.FullName + `/pull/1396"}`),
	})

	var first AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex", RequestID: "first-missing-branch", Instruction: "Verify the shortcut implementation.",
	}, http.StatusAccepted, &first)
	pollRequest := PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}
	poll, err := f.store.PollAgent(t.Context(), first.AgentID, pollRequest)
	if err != nil || poll.Assignment == nil || poll.Assignment.Worktree.Source != "pull_request_fallback:1396" {
		t.Fatalf("assignment did not include its sole linked PR: %+v, %v", poll.Assignment, err)
	}
	f.post(t, "/v1/agents/"+first.AgentID+"/events", AgentEventRequest{
		AssignmentID: first.AssignmentID, TaskID: first.TaskID, DaemonID: pollRequest.DaemonID, DeviceID: pollRequest.DeviceID, RuntimeID: pollRequest.RuntimeID,
		State: AssignmentFailed, EventType: EventAssignmentFailed,
		Metadata: map[string]string{assignment.ExecutionDiagnosticKey: "workspace_preparation_failed", assignment.ExecutionNextActionKey: "check_workspace"},
	}, http.StatusOK, nil)

	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	f.post(t, path, CreateAIAgentTaskThreadMessageRequest{Body: "PR #42", RequestID: "wrong-pr"}, http.StatusBadRequest, nil)
	f.token = "native-teammate-token"
	f.post(t, path, CreateAIAgentTaskThreadMessageRequest{Body: "PR #1396", RequestID: "other-user"}, http.StatusForbidden, nil)
	f.token = "native-owner-token"
	input := CreateAIAgentTaskThreadMessageRequest{Body: "<p>PR #1396</p>", RequestID: "retry-workspace"}
	var next, replay AIAgentTaskActionResponse
	f.post(t, path, input, http.StatusAccepted, &next)
	f.post(t, path, input, http.StatusAccepted, &replay)
	f.post(t, path, CreateAIAgentTaskThreadMessageRequest{Body: "branch: staging", RequestID: "second-retry"}, http.StatusConflict, nil)
	poll, err = f.store.PollAgent(t.Context(), first.AgentID, pollRequest)
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID || next.AssignmentID != replay.AssignmentID ||
		poll.Assignment.ResumeSessionID != "" || poll.Assignment.Worktree.BranchName != "refs/pull/1396/head" ||
		!strings.Contains(poll.Assignment.Prompt, "### New User Instruction\nVerify the shortcut implementation.") {
		t.Fatalf("workspace retry lost its checkout or original instruction: %+v, %v", poll.Assignment, err)
	}
	history, err := f.client.ListAIAgentTaskThreadHistory(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, "task-new")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range history.Threads {
		if row.ThreadID == next.ThreadID && row.ConversationID == first.ThreadID && row.ParentThreadID == first.ThreadID {
			return
		}
	}
	t.Fatal("workspace retry did not stay in the original Riido conversation")
}
