package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPMembershipRemovalRejectsGenerationChangedAtDynamoCommit(t *testing.T) {
	const taskID, agentID, token = "task-membership-commit-fence", "agent-public-openclaw", "ai-agent-token"
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
	f.mu.Lock()
	f.beforeScopedCommit = func(snapshot AIAgentClientSnapshot) {
		if snapshot.TaskAgentMemberships[taskID][agentID].RemovedAt != nil {
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
		t.Fatal("removal did not reach its guarded commit")
	}
	post := httptest.NewRequest(http.MethodPost, base+"/tasks/"+taskID+"/assignment", strings.NewReader(`{"agent_id":"`+agentID+`"}`))
	post.Header.Set(aiAgentTokenHeader, token)
	posted := httptest.NewRecorder()
	publisher.ServeHTTP(posted, post)
	close(resume)
	removed := <-result
	if posted.Code != http.StatusAccepted || removed.Code != http.StatusConflict {
		t.Fatalf("publication=%d removal=%d; want accepted/conflict", posted.Code, removed.Code)
	}
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	profiles, err := clientA.ListWorkspaceAssignedAgentProfiles(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	members := profiles.TaskAgentParticipants[taskID].Agents
	if len(members) != 1 || members[0].LatestAssignmentID == first.AssignmentID {
		t.Fatalf("failed commit leaked tombstone or stale thread: %+v", members)
	}
	projection, found, err := storeA.operationStore.(AssignmentProjectionReader).LoadAssignmentProjection(ctx, members[0].LatestAssignmentID)
	if err != nil || !found || projection.Assignment.State != AssignmentQueued {
		t.Fatalf("unobserved execution was cancelled: %+v err=%v", projection, err)
	}
}
