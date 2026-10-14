package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type durableBeforePublishGate struct {
	*Store
	entered chan Assignment
	resume  chan struct{}
	armed   atomic.Bool
}

func (s *durableBeforePublishGate) waitPublication(ctx context.Context, a Assignment, err error) (Assignment, error) {
	if err != nil || !s.armed.Load() {
		return a, err
	}
	s.entered <- a
	select {
	case <-s.resume:
		return a, nil
	case <-ctx.Done():
		return Assignment{}, ctx.Err()
	}
}
func (s *durableBeforePublishGate) AssignTaskReplacement(ctx context.Context, taskID string, req AssignRequest) (Assignment, error) {
	a, err := s.Store.AssignTaskReplacement(ctx, taskID, req)
	return s.waitPublication(ctx, a, err)
}
func (s *durableBeforePublishGate) AssignTaskAdditive(ctx context.Context, taskID string, req AssignRequest) (Assignment, error) {
	a, err := s.Store.AssignTaskAdditive(ctx, taskID, req)
	return s.waitPublication(ctx, a, err)
}
func (s *durableBeforePublishGate) AssignTask(ctx context.Context, taskID string, req AssignRequest) (Assignment, error) {
	a, err := s.Store.AssignTask(ctx, taskID, req)
	return s.waitPublication(ctx, a, err)
}
func TestHTTPMembershipRemovalWaitsForDurablePublication(t *testing.T) {
	for _, mode := range []string{"replacement", "additive", "reply"} {
		for _, state := range []string{"queued", "completed"} {
			t.Run(mode+"/"+state, func(t *testing.T) { checkMembershipRemovalDurablePublication(t, mode, state) })
		}
	}
}
func checkMembershipRemovalDurablePublication(t *testing.T, mode, state string) {
	const taskID, agentID, token = "task-review-durable-before-publish", "agent-public-openclaw", "ai-agent-token"
	client := NewDevelopmentAIAgentClientStore()
	assignments := &durableBeforePublishGate{Store: NewStoreWithConfig(StoreConfig{AgentRegistry: client}), entered: make(chan Assignment, 1), resume: make(chan struct{})}
	defer assignments.Close()
	handler := NewServer(ServerConfig{AIAgentClient: client, Assignment: assignments, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	first := createProviderMultiAssignment(t, handler, base, token, taskID, agentID)
	stopReq := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments/"+agentID+"/stop", strings.NewReader(`{"assignment_id":"`+first.AssignmentID+`"}`))
	stopReq.Header.Set(aiAgentTokenHeader, token)
	stopped := httptest.NewRecorder()
	handler.ServeHTTP(stopped, stopReq)
	if stopped.Code != http.StatusAccepted {
		t.Fatalf("stop=%d %s", stopped.Code, stopped.Body.String())
	}
	assignments.armed.Store(true)
	postPath, postBody := base+"/tasks/"+taskID+"/assignment", `{"agent_id":"`+agentID+`"}`
	if mode == "additive" {
		postPath = base + "/tasks/" + taskID + "/agent-assignments"
	}
	if mode == "reply" {
		postPath = base + "/tasks/" + taskID + "/threads/" + first.ThreadID + "/messages"
		postBody = `{"body":"continue this task"}`
	}

	post := httptest.NewRequest(http.MethodPost, postPath, strings.NewReader(postBody))
	post.Header.Set(aiAgentTokenHeader, token)
	postResult := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() { handler.ServeHTTP(postResult, post); close(finished) }()
	var next Assignment
	select {
	case next = <-assignments.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("did not reach durable-before-publish pause")
	}

	if state == "completed" {
		poll, err := assignments.Store.PollAgent(context.Background(), agentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
		if err != nil || poll.Assignment == nil || poll.Assignment.ID != next.ID {
			t.Fatalf("unpublished poll=%+v err=%v", poll, err)
		}
		mustRecordActorEvent(t, assignments.Store, context.Background(), agentID, AgentEventRequest{AssignmentID: next.ID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", State: AssignmentRunning, EventType: EventAssignmentRunning, Message: "started before publication"})
		mustRecordActorEvent(t, assignments.Store, context.Background(), agentID, AgentEventRequest{AssignmentID: next.ID, DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared", State: AssignmentCompleted, EventType: EventAssignmentCompleted, Message: "done before publication"})
	}
	removedResult := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		removedResult <- serveUnassignTaskBoundary(handler, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, `{"assignment_id":"`+first.AssignmentID+`","reason":"participant_removed"}`)
	}()
	select {
	case removed := <-removedResult:
		close(assignments.resume)
		<-finished
		t.Fatalf("removal crossed unpublished durable execution: status=%d body=%s", removed.Code, removed.Body.String())
	case <-time.After(20 * time.Millisecond):
	}
	close(assignments.resume)
	<-finished
	removed := <-removedResult
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	after := membershipAgentIDs(t, client, p, taskID)
	poll, err := assignments.PollAgent(context.Background(), agentID, PollRequest{DaemonID: "daemon-shared-studio", DeviceID: "device-shared-studio", RuntimeID: "runtime-openclaw-shared"})
	if postResult.Code != http.StatusAccepted || len(after) != 1 || after[0] != agentID {
		t.Fatalf("post=%d members=%v", postResult.Code, after)
	}
	if state == "queued" && (err != nil || poll.Action != PollStart || poll.Assignment == nil || poll.Assignment.ID != next.ID) {
		t.Fatalf("new execution not preserved: poll=%+v err=%v", poll, err)
	}
	if removed.Code != http.StatusConflict {
		t.Fatalf("unpublished durable execution crossed removal: DELETE=%d; expected409", removed.Code)
	}
}
