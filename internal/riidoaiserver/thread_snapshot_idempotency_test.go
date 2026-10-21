package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestThreadSnapshotRequestRetryKeepsFirstExecutionAndHistory(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	req := CreateAIAgentTaskThreadMessageRequest{Body: "First followup", RequestID: "same-click", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: "<p>first snapshot</p>", ContentFormat: "html"}}
	var next, replay AIAgentTaskActionResponse
	f.post(t, path, req, http.StatusAccepted, &next)
	req.Body, req.DocumentSnapshot.Content = "changed retry", "<p>changed snapshot</p>"
	f.post(t, path, req, http.StatusAccepted, &replay)
	if next.AssignmentID != replay.AssignmentID || next.ThreadID != replay.ThreadID {
		t.Fatal("retry created another execution/thread")
	}
	projection, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
	if err != nil || !found || !strings.Contains(projection.Assignment.Prompt, "first snapshot") || strings.Contains(projection.Assignment.Prompt, "changed snapshot") || projection.Assignment.ResumeSessionID != "native-http-session" {
		t.Fatalf("first snapshot/native resume lost: %+v err=%v", projection, err)
	}
	history, err := f.client.ListAIAgentTaskThreadHistory(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, "task-new")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, thread := range history.Threads {
		for _, message := range thread.Messages {
			if message.Role == AIAgentTaskThreadMessageRoleUser {
				count++
			}
		}
	}
	if count != 1 {
		t.Fatalf("user history count=%d", count)
	}
}

func TestActiveThreadSnapshotRequestKeepsExistingExecution(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	var first, next, replay AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/agent-assignments", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", RequestID: "initial"}, http.StatusAccepted, &first)
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	req := CreateAIAgentTaskThreadMessageRequest{Body: "Follow active task", RequestID: "active-click"}
	f.post(t, path, req, http.StatusAccepted, &next)
	f.post(t, path, req, http.StatusAccepted, &replay)
	if first.AssignmentID != next.AssignmentID || next.AssignmentID != replay.AssignmentID || first.ThreadID != next.ThreadID {
		t.Fatal("active followup changed legacy execution semantics")
	}
}
