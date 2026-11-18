package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestHTTPNativeFollowupRetainsWorktreeAfterNewerPullIsLinked(t *testing.T) {
	for _, terminal := range []AssignmentState{AssignmentCompleted, AssignmentFailed, AssignmentCancelled} {
		t.Run(string(terminal), func(t *testing.T) {
			f := automaticCheckoutHTTPFixture(t)
			first := f.finishFirst(t, terminal)
			previous, _, err := f.store.LoadAssignmentProjection(t.Context(), first.AssignmentID)
			if err != nil {
				t.Fatal(err)
			}
			repo := f.taskContext.contextSnapshot.Repositories[0].FullName
			f.taskContext.contextSnapshot.linkedPullRequests = append(f.taskContext.contextSnapshot.linkedPullRequests,
				taskCheckoutPullRequests([]json.RawMessage{json.RawMessage(`{"number":23,"updatedAt":"2026-11-13T01:00:00Z","htmlUrl":"https://github.com/` + repo + `/pull/23"}`)})...)
			var next AIAgentTaskActionResponse
			path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
			request := CreateAIAgentTaskThreadMessageRequest{Body: "Continue reviewing PR #19 in this conversation.", RequestID: "reply-after-new-pull"}
			f.post(t, path, request, http.StatusAccepted, &next)
			var replay AIAgentTaskActionResponse
			f.post(t, path, request, http.StatusAccepted, &replay)
			if replay.AssignmentID != next.AssignmentID {
				t.Fatal("retry created a second execution")
			}
			poll, err := f.store.PollAgent(t.Context(), first.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
			if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.AssignmentID || poll.Assignment.ResumeSessionID != "native-http-session" || !reflect.DeepEqual(poll.Assignment.Worktree, previous.Assignment.Worktree) {
				t.Fatal("reply did not reach daemon with its original session/worktree", err)
			}
		})
	}
}
