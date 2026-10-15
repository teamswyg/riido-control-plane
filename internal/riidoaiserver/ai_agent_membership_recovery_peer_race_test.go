package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPFailedPublicationRecoveryRetriesUnrelatedMembershipCommit(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-recovery-peer-commit")
	first := s.stoppedAssignment(t)
	clientB, storeB := newMembershipDynamoReplica(t, s.f)
	clientC, storeC := newMembershipDynamoReplica(t, s.f)
	publisherB, publisherC := s.replicaHandler(clientB, storeB), s.replicaHandler(clientC, storeC)
	entered, resume := make(chan struct{}), make(chan struct{})
	s.f.mu.Lock()
	s.f.rejectScopedPublication = true
	s.f.beforePublicationRecovery = func() {
		s.f.mu.Lock()
		s.f.rejectScopedPublication = false
		s.f.beforePublicationRecovery = nil
		s.f.mu.Unlock()
		close(entered)
		<-resume
	}
	s.f.mu.Unlock()
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() { result <- s.post(publisherB, "/assignment", `{"agent_id":"`+s.agentID+`"}`) }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		close(resume)
		t.Fatal("recovery did not reach cancellation commit")
	}
	peer := createProviderMultiAssignment(t, publisherC, s.base, s.token, s.taskID, "agent-owned-codex")
	close(resume)
	if (<-result).Code == http.StatusAccepted {
		t.Fatal("failed publication was accepted")
	}
	removed := serveUnassignTaskBoundary(s.handler, s.base+"/tasks/"+s.taskID+"/agent-assignments/"+s.agentID, s.token, `{"assignment_id":"`+first.AssignmentID+`"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("unrelated member commit left failed publication blocked: remove=%d %s", removed.Code, removed.Body.String())
	}
	members := s.participants(t)
	if len(members) != 1 || members[0].AgentID != peer.AgentID {
		t.Fatalf("recovery or removal lost peer: %+v", members)
	}
}
