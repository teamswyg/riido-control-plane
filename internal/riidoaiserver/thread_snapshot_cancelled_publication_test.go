package riidoaiserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

type snapshotCancellingPublicationStore struct {
	*Store
	claim bool
}

func (s snapshotCancellingPublicationStore) recoverFailedTaskAgentPublication(ctx context.Context, assignment Assignment) error {
	if s.claim {
		poll, err := s.PollAgent(ctx, assignment.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
		if err != nil {
			return err
		}
		if poll.Assignment == nil || poll.Assignment.ID != assignment.ID {
			return fmt.Errorf("expected assignment %s to be claimed", assignment.ID)
		}
	}
	_, err := s.CancelAssignment(ctx, assignment.TaskID, CancelAssignmentRequest{AgentID: assignment.AgentID, AssignmentID: assignment.ID, Reason: "client publication failed"})
	return err
}

func TestThreadSnapshotCancelledPublicationRetryDoesNotReportAccepted(t *testing.T) {
	for _, claim := range []bool{false, true} {
		t.Run(fmt.Sprint(claim), func(t *testing.T) { checkThreadSnapshotCancelledPublicationRetry(t, claim) })
	}
}

func checkThreadSnapshotCancelledPublicationRetry(t *testing.T, claim bool) {
	t.Helper()
	expected := AssignmentCancelled
	if claim {
		expected = AssignmentCancelling
	}
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	snapshot := aiAgentTaskContextHTTPFixture()
	snapshot.Component.ID = "task-new"
	publisher := &snapshotFailOnceThreadPublisher{DevelopmentAIAgentClientStore: f.client, fail: true}
	f.handler = NewServer(ServerConfig{AIAgentClient: publisher, Assignment: snapshotCancellingPublicationStore{Store: f.store, claim: claim}, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: snapshot}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	f.token = "ai-agent-token"
	path := f.base + "/tasks/task-new/threads/" + first.ThreadID + "/messages"
	req := CreateAIAgentTaskThreadMessageRequest{Body: "Retry cancelled publication", RequestID: "cancelled-click", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: "<p>original body</p>", ContentFormat: "html"}}
	f.post(t, path, req, http.StatusBadRequest, nil)
	id := identifiedAssignmentID("task-new", AssignRequest{AgentID: "agent-owned-codex", CreatedBy: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID, RequestID: threadMessageRequestID(first.ThreadID, req.RequestID)})
	before, found, err := f.store.LoadAssignmentProjection(t.Context(), id)
	if err != nil || !found || before.Assignment.State != expected {
		t.Fatalf("publication recovery did not cancel execution: %+v %v", before, err)
	}
	metricsBefore, err := f.store.Metrics(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	f.post(t, path, req, http.StatusServiceUnavailable, nil)
	after, found, err := f.store.LoadAssignmentProjection(t.Context(), id)
	if err != nil || !found || after.Assignment.State != expected || before.Assignment.Prompt != after.Assignment.Prompt {
		t.Fatalf("retry rewrote cancelled execution: %+v %v", after, err)
	}
	metricsAfter, err := f.store.Metrics(t.Context())
	if err != nil || metricsBefore.AssignmentsTotal != metricsAfter.AssignmentsTotal || publisher.calls != 1 {
		t.Fatalf("retry created execution or published cancelled work: before=%+v after=%+v calls=%d err=%v", metricsBefore, metricsAfter, publisher.calls, err)
	}
}
