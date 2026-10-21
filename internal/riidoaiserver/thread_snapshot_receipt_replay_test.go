package riidoaiserver

import (
	"errors"
	"net/http"
	"testing"
)

func TestThreadSnapshotReceiptReplayDoesNotReadTaskContextAgain(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	snapshot := aiAgentTaskContextHTTPFixture()
	snapshot.Component.ID = "task-new"
	reader := &assignmentHTTPRequestTaskContextReader{contextSnapshot: snapshot}
	f.handler = NewServer(ServerConfig{AIAgentClient: f.client, Assignment: f.store, TaskContext: reader, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	f.token = "ai-agent-token"
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	req := CreateAIAgentTaskThreadMessageRequest{Body: "Reply once", RequestID: "lost-response", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: "<p>click body</p>", ContentFormat: "html"}}
	var next, replay AIAgentTaskActionResponse
	f.post(t, path, req, http.StatusAccepted, &next)
	reader.err = errors.New("temporary context network failure")
	f.post(t, path, req, http.StatusAccepted, &replay)
	if len(reader.requests) != 1 || replay.AssignmentID != next.AssignmentID || replay.ThreadID != next.ThreadID {
		t.Fatal("receipt replay reread context or created another execution")
	}
	// A new request with a failed context read remains retryable, never a 400 that
	// could cause the client to discard the original execution identity.
	req.RequestID = "new-click"
	f.post(t, path, req, http.StatusServiceUnavailable, nil)
}
