package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Separate HTTP servers must not let removal commit across a durable execution
// whose task-thread publication is still in flight on the other server.
func TestHTTPMembershipRemovalAcrossServersPreservesUnpublishedExecution(t *testing.T) {
	const taskID, agentID, token = "task-cross-server-removal", "agent-public-openclaw", "ai-agent-token"
	client := NewDevelopmentAIAgentClientStore()
	assignments := &durableBeforePublishGate{Store: NewStoreWithConfig(StoreConfig{AgentRegistry: client}), entered: make(chan Assignment, 1), resume: make(chan struct{})}
	defer assignments.Close()
	config := ServerConfig{AIAgentClient: client, Assignment: assignments, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}
	publisher, remover := NewServer(config).Handler(), NewServer(config).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	first := createProviderMultiAssignment(t, publisher, base, token, taskID, agentID)
	stop := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments/"+agentID+"/stop", strings.NewReader(`{"assignment_id":"`+first.AssignmentID+`"}`))
	stop.Header.Set(aiAgentTokenHeader, token)
	stopped := httptest.NewRecorder()
	publisher.ServeHTTP(stopped, stop)
	if stopped.Code != http.StatusAccepted {
		t.Fatalf("initial stop=%d %s", stopped.Code, stopped.Body.String())
	}
	assignments.armed.Store(true)
	post := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/assignment", strings.NewReader(`{"agent_id":"`+agentID+`"}`))
	post.Header.Set(aiAgentTokenHeader, token)
	posted := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() { publisher.ServeHTTP(posted, post); close(finished) }()
	select {
	case <-assignments.entered:
	case <-time.After(3 * time.Second):
		close(assignments.resume)
		<-finished
		t.Fatal("did not reach durable-before-publication boundary")
	}
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		result <- serveUnassignTaskBoundary(remover, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, `{"assignment_id":"`+first.AssignmentID+`","reason":"participant_removed"}`)
	}()
	var removed *httptest.ResponseRecorder
	select {
	case removed = <-result:
	case <-time.After(20 * time.Millisecond):
	}
	close(assignments.resume)
	<-finished
	if removed == nil {
		removed = <-result
	}
	if posted.Code != http.StatusAccepted {
		t.Fatalf("new execution publication=%d %s", posted.Code, posted.Body.String())
	}
	if removed.Code != http.StatusConflict {
		t.Fatalf("other server removed membership across unpublished durable execution: DELETE=%d, expected 409", removed.Code)
	}
}
