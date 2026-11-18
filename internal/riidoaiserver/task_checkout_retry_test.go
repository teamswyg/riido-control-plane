package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPWorkspaceRetryUsesLatestPRUnlessUserSelectsAnother(t *testing.T) {
	for _, test := range []struct{ body, branch string }{
		{"Continue reviewing.", "refs/pull/1397/head"},
		{"PR #1396", "refs/pull/1396/head"},
		{"branch: staging", "staging"},
	} {
		t.Run(test.body, func(t *testing.T) {
			f, first, request := workspaceRetryHTTPFixture(t, false)
			repo := f.taskContext.contextSnapshot.Repositories[0].FullName
			f.taskContext.contextSnapshot.linkedPullRequests = append(f.taskContext.contextSnapshot.linkedPullRequests,
				taskCheckoutPullRequests([]json.RawMessage{json.RawMessage(`{"number":1397,"updatedAt":"2026-11-12T01:00:00Z","htmlUrl":"https://github.com/` + repo + `/pull/1397"}`)})...)
			var result AIAgentTaskActionResponse
			f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages",
				CreateAIAgentTaskThreadMessageRequest{Body: test.body, RequestID: "updated-pr-retry"}, http.StatusAccepted, &result)
			poll, err := f.store.PollAgent(t.Context(), first.AgentID, request)
			if err != nil || poll.Assignment == nil || poll.Assignment.ID != result.AssignmentID || poll.Assignment.Worktree.BranchName != test.branch {
				t.Fatalf("workspace retry did not honor refreshed checkout selection: %+v, %v", poll.Assignment, err)
			}
		})
	}
}
