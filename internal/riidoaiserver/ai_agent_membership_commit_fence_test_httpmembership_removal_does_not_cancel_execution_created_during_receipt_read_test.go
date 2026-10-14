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

func TestHTTPMembershipRemovalDoesNotCancelExecutionCreatedDuringReceiptRead(t *testing.T) {
	const taskID, agentID, token = "task-membership-receipt-read-race", "agent-public-openclaw", "ai-agent-token"
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
	publisher := NewServer(config(clientB, storeB)).Handler()
	entered, resume := make(chan struct{}), make(chan struct{})
	var armed atomic.Bool
	armed.Store(true)
	f.mu.Lock()
	f.beforeReceiptQuery = func() {
		if armed.Swap(false) {
			close(entered)
			<-resume
		}
	}
	f.mu.Unlock()
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		result <- serveUnassignTaskBoundary(remover, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, `{"assignment_id":"`+first.AssignmentID+`"}`)
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		close(resume)
		t.Fatal("removal did not pause between marker and receipts")
	}
	posted := createProviderMultiAssignment(t, publisher, base, token, taskID, agentID)
	close(resume)
	removed := <-result
	if removed.Code != http.StatusConflict {
		t.Fatalf("remove=%d want409", removed.Code)
	}
	projection, found, err := storeA.operationStore.(AssignmentProjectionReader).LoadAssignmentProjection(ctx, posted.AssignmentID)
	if err != nil || !found || projection.Assignment.State != AssignmentQueued {
		t.Fatalf("conflicting removal cancelled the unobserved execution: %+v err=%v", projection, err)
	}
}
