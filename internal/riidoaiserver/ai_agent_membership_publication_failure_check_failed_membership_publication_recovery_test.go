package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func checkFailedMembershipPublicationRecovery(t *testing.T, state string) {
	t.Helper()
	const taskID, agentID, token = "task-publication-failure-recovery", "agent-public-openclaw", "ai-agent-token"
	ctx := context.Background()
	f := newMembershipSharedDynamoFixture(t)
	clientA, storeA := newMembershipDynamoReplica(t, f)
	config := func(client AIAgentClientStore, store AssignmentStore) ServerConfig {
		return ServerConfig{AIAgentClient: client, Assignment: store, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}
	}
	remover := NewServer(config(clientA, storeA)).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	first := createProviderMultiAssignment(t, remover, base, token, taskID, agentID)
	stop := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments/"+agentID+"/stop", strings.NewReader(`{"assignment_id":"`+first.AssignmentID+`"}`))
	stop.Header.Set(aiAgentTokenHeader, token)
	stopped := httptest.NewRecorder()
	remover.ServeHTTP(stopped, stop)
	if stopped.Code != http.StatusAccepted {
		t.Fatalf("stop=%d", stopped.Code)
	}
	clientB, storeB := newMembershipDynamoReplica(t, f)
	paused := &durableBeforePublishGate{Store: storeB, entered: make(chan Assignment, 1), resume: make(chan struct{})}
	paused.armed.Store(true)
	publisher := NewServer(config(clientB, paused)).Handler()
	f.mu.Lock()
	f.rejectScopedPublication = true
	f.mu.Unlock()
	post := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/assignment", strings.NewReader(`{"agent_id":"`+agentID+`"}`))
	post.Header.Set(aiAgentTokenHeader, token)
	posted := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() { publisher.ServeHTTP(posted, post); close(finished) }()
	var pending Assignment
	select {
	case pending = <-paused.entered:
	case <-time.After(3 * time.Second):
		close(paused.resume)
		<-finished
		t.Fatal("publication did not pause after durable creation")
	}
	if state != "queued" {
		advanceFailedPublicationExecution(t, storeB, pending, state)
	}
	before, _, err := storeB.operationStore.(AssignmentProjectionReader).LoadAssignmentProjection(ctx, pending.ID)
	if err != nil {
		close(paused.resume)
		<-finished
		t.Fatal(err)
	}
	close(paused.resume)
	<-finished
	f.mu.Lock()
	f.rejectScopedPublication = false
	f.mu.Unlock()
	if posted.Code == http.StatusAccepted {
		t.Fatal("injected publication failure was accepted")
	}
	after, found, err := storeB.operationStore.(AssignmentProjectionReader).LoadAssignmentProjection(ctx, pending.ID)
	if err != nil || !found {
		t.Fatalf("load failure recovery: %v", err)
	}
	want := AssignmentCancelled
	if state == "running" {
		want = AssignmentCancelling
	}
	if state == "completed" || state == "failed" {
		want = before.Assignment.State
		if !reflect.DeepEqual(after.Assignment, before.Assignment) || after.LastEventSeq != before.LastEventSeq {
			t.Fatal("failure recovery rewrote terminal execution data")
		}
	}
	if after.Assignment.State != want {
		t.Fatalf("failed execution state=%s want=%s", after.Assignment.State, want)
	}
	removed := serveUnassignTaskBoundary(remover, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, `{"assignment_id":"`+first.AssignmentID+`"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("failed publication permanently blocks explicit removal: DELETE=%d %s", removed.Code, removed.Body.String())
	}
	_, active, err := storeA.loadTaskAgentDurableExecutions(ctx, taskID, agentID)
	remaining := 0
	if state == "running" {
		remaining = 1
	}
	if err != nil || len(active) != remaining {
		t.Fatalf("failed publication execution survived removal unexpectedly: active=%d err=%v", len(active), err)
	}
}
