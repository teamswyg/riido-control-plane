package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPMembershipScopedRemovalPreservesPeerPublishedAfterSnapshotLoad(t *testing.T) {
	const taskID, agentID, token = "task-scoped-peer-history", "agent-public-openclaw", "ai-agent-token"
	ctx := context.Background()
	f := newMembershipSharedDynamoFixture(t)
	clientA, storeA := newMembershipDynamoReplica(t, f)
	config := func(client AIAgentClientStore, store AssignmentStore) ServerConfig {
		return ServerConfig{AIAgentClient: client, Assignment: store, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}
	}
	remover := NewServer(config(clientA, storeA)).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	first := createProviderMultiAssignment(t, remover, base, token, taskID, agentID)
	clientB, storeB := newMembershipDynamoReplica(t, f)
	publisher := NewServer(config(clientB, storeB)).Handler()
	entered, resume := make(chan struct{}), make(chan struct{})
	var armed atomic.Bool
	armed.Store(true)
	f.mu.Lock()
	f.beforeOwnedSnapshotRead = func() {
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
		t.Fatal("scoped removal did not pause before fresh membership read")
	}
	peer := createProviderMultiAssignment(t, publisher, base, token, taskID, "agent-owned-codex")
	close(resume)
	removed := <-result
	if removed.Code != http.StatusAccepted {
		t.Fatalf("remove=%d %s", removed.Code, removed.Body.String())
	}
	snapshot, found, err := clientA.snapshotStore.LoadAIAgentClientSnapshot(ctx)
	if err != nil || !found {
		t.Fatalf("load committed snapshot: %v", err)
	}
	sawPeer := false
	for _, thread := range snapshot.TaskThreads[taskID] {
		if thread.AssignmentID == peer.AssignmentID {
			sawPeer = true
		}
	}
	if !sawPeer {
		t.Fatal("scoped removal dropped another Agent's newly published thread")
	}
	member := snapshot.TaskAgentMemberships[taskID]["agent-owned-codex"]
	if member.RemovedAt != nil || member.AgentID != "agent-owned-codex" {
		t.Fatalf("peer membership changed: %+v", member)
	}
}
