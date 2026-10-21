package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type snapshotFailOnceThreadPublisher struct {
	*DevelopmentAIAgentClientStore
	fail bool
}

func (s *snapshotFailOnceThreadPublisher) CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error) {
	if s.fail {
		s.fail = false
		return AIAgentTaskActionResponse{}, errors.New("injected thread publication failure")
	}
	return s.DevelopmentAIAgentClientStore.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, req)
}

func TestThreadSnapshotRetryAfterPublicationFailureDoesNotExecuteTwice(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	contextSnapshot := aiAgentTaskContextHTTPFixture()
	contextSnapshot.Component.ID = "task-new"
	publisher := &snapshotFailOnceThreadPublisher{DevelopmentAIAgentClientStore: f.client, fail: true}
	f.handler = NewServer(ServerConfig{AIAgentClient: publisher, Assignment: f.store, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: contextSnapshot}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	f.token = "ai-agent-token"
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	req := CreateAIAgentTaskThreadMessageRequest{Body: "Retry publication", RequestID: "publish-click", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: "<p>original body</p>", ContentFormat: "html"}}
	f.post(t, path, req, http.StatusBadRequest, nil)
	id := identifiedAssignmentID("task-new", AssignRequest{AgentID: "agent-owned-codex", CreatedBy: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID, RequestID: threadMessageRequestID(first.ThreadID, req.RequestID)})
	before, found, err := f.store.LoadAssignmentProjection(t.Context(), id)
	if err != nil || !found {
		t.Fatalf("initial execution missing: %v", err)
	}
	var replay AIAgentTaskActionResponse
	f.post(t, path, req, http.StatusAccepted, &replay)
	after, found, err := f.store.LoadAssignmentProjection(t.Context(), id)
	if err != nil || !found || replay.AssignmentID != id || before.Assignment.Prompt != after.Assignment.Prompt || before.Assignment.State != after.Assignment.State {
		t.Fatal("failed publication retry recreated or rewrote execution")
	}
}
