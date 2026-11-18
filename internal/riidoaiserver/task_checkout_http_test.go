package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPAssignmentCheckoutAutomaticAndUserSelection(t *testing.T) {
	for _, test := range []struct{ instruction, branch string }{
		{"Review this task.", "refs/pull/19/head"},
		{"PR #21\nReview the older PR.", "refs/pull/21/head"},
		{"<p>branch: staging</p><p>Review on staging.</p>", "staging"},
		{"브랜치: 작업/수정\nContinue.", "작업/수정"},
	} {
		t.Run(test.instruction, func(t *testing.T) {
			f := automaticCheckoutHTTPFixture(t)
			var result AIAgentTaskActionResponse
			f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{
				AgentID: "agent-owned-codex", RequestID: "checkout", Instruction: test.instruction,
			}, http.StatusAccepted, &result)
			poll, err := f.store.PollAgent(t.Context(), result.AgentID, PollRequest{
				DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev",
			})
			if err != nil || poll.Assignment == nil || poll.Assignment.Worktree == nil || poll.Assignment.Worktree.BranchName != test.branch {
				t.Fatalf("daemon checkout did not honor selection: %+v, %v", poll.Assignment, err)
			}
		})
	}
}

func TestHTTPAssignmentCheckoutInvalidSelectionDoesNotQueue(t *testing.T) {
	for _, instruction := range []string{"PR #404", "branch: --invalid", "branch: refs/pull/19/head"} {
		t.Run(instruction, func(t *testing.T) {
			f := automaticCheckoutHTTPFixture(t)
			f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{
				AgentID: "agent-owned-codex", RequestID: "invalid-checkout", Instruction: instruction,
			}, http.StatusBadRequest, nil)
			poll, err := f.store.PollAgent(t.Context(), "agent-owned-codex", PollRequest{
				DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev",
			})
			if err != nil || poll.Assignment != nil {
				t.Fatalf("invalid checkout queued a contextless assignment: %+v, %v", poll.Assignment, err)
			}
		})
	}
}

func automaticCheckoutHTTPFixture(t *testing.T) *nativeResumeHTTPFixture {
	t.Helper()
	f := newNativeResumeHTTPFixture(t)
	snapshot := &f.taskContext.contextSnapshot
	snapshot.Component.KeyNumber, snapshot.Component.BranchName = "RIID-6119", "RIID-6119-generated-title"
	repo := snapshot.Repositories[0].FullName
	snapshot.linkedPullRequests = taskCheckoutPullRequests([]json.RawMessage{
		json.RawMessage(`{"number":21,"state":"OPEN","updatedAt":"2026-11-11T01:00:00Z","htmlUrl":"https://github.com/` + repo + `/pull/21"}`),
		json.RawMessage(`{"number":19,"state":"MERGED","updatedAt":"2026-11-12T01:00:00Z","htmlUrl":"https://github.com/` + repo + `/pull/19"}`),
	})
	return f
}
