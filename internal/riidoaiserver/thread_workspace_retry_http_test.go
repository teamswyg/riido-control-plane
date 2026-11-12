package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestHTTPWorkspaceRetryPreservesConversationAndInstruction(t *testing.T) {
	t.Run("linked PR selection", func(t *testing.T) {
		testHTTPWorkspaceRetryReply(t, "<p>PR #1396</p>", "Verify the shortcut implementation.", "refs/pull/1396/head")
	})
	t.Run("ordinary conversation reply", func(t *testing.T) {
		testHTTPWorkspaceRetryReply(t, "<p>Check the Chrome shortcut as well.</p>", "Check the Chrome shortcut as well.", "refs/pull/1396/head")
	})
	t.Run("ordinary reply after branch settings changed", func(t *testing.T) {
		testHTTPWorkspaceRetryReply(t, "<p>Use the updated settings and continue.</p>", "Use the updated settings and continue.", "staging")
	})
}

func testHTTPWorkspaceRetryReply(t *testing.T, body, instruction, expectedRef string) {
	t.Helper()
	f, first, pollRequest := workspaceRetryHTTPFixture(t)

	if expectedRef == "staging" {
		f.taskContext.contextSnapshot.Component.BranchName = expectedRef
	}
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	f.post(t, path, CreateAIAgentTaskThreadMessageRequest{Body: "PR #42", RequestID: "wrong-pr"}, http.StatusBadRequest, nil)
	f.token = "native-teammate-token"
	f.post(t, path, CreateAIAgentTaskThreadMessageRequest{Body: "PR #1396", RequestID: "other-user"}, http.StatusForbidden, nil)
	f.token = "native-owner-token"
	input := CreateAIAgentTaskThreadMessageRequest{Body: body, RequestID: "retry-workspace"}
	var next, replay AIAgentTaskActionResponse
	f.post(t, path, input, http.StatusAccepted, &next)
	f.post(t, path, input, http.StatusAccepted, &replay)
	f.post(t, path, CreateAIAgentTaskThreadMessageRequest{Body: "branch: staging", RequestID: "second-retry"}, http.StatusConflict, nil)
	poll, err := f.store.PollAgent(t.Context(), first.AgentID, pollRequest)
	if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID || next.AssignmentID != replay.AssignmentID ||
		poll.Assignment.ResumeSessionID != "" || poll.Assignment.Worktree.BranchName != expectedRef ||
		!strings.Contains(poll.Assignment.Prompt, "### New User Instruction\n"+instruction) {
		t.Fatalf("workspace retry lost its checkout or original instruction: %+v, %v", poll.Assignment, err)
	}
	history, err := f.client.ListAIAgentTaskThreadHistory(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, "task-new")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range history.Threads {
		if row.ThreadID == next.ThreadID && row.ConversationID == first.ThreadID && row.ParentThreadID == first.ThreadID {
			for _, message := range row.Messages {
				if message.Role == AIAgentTaskThreadMessageRoleUser && message.Body == body {
					return
				}
			}
			t.Fatalf("retry reply was not saved in conversation history: %+v", row.Messages)
		}
	}
	t.Fatal("workspace retry did not stay in the original Riido conversation")
}
