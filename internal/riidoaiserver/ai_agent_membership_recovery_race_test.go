package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPFailedPublicationRecoveryCancelsOnlySupersededUnpublishedExecution(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-recovery-new-execution")
	s.stoppedAssignment(t)
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
		t.Fatal("failure recovery did not reach guarded cancellation")
	}
	newer := createProviderMultiAssignment(t, publisherC, s.base, s.token, s.taskID, s.agentID)
	close(resume)
	if (<-result).Code == http.StatusAccepted {
		t.Fatal("failed publication was accepted")
	}
	members := s.participants(t)
	if len(members) != 1 || members[0].LatestAssignmentID != newer.AssignmentID {
		t.Fatalf("recovery rewound latest execution: %+v", members)
	}
	_, active, err := s.store.loadTaskAgentDurableExecutions(context.Background(), s.taskID, s.agentID)
	if err != nil || len(active) != 1 {
		t.Fatalf("failed recovery left a superseded execution runnable: active=%d err=%v", len(active), err)
	}
	for _, projection := range active {
		if projection.Assignment.State != AssignmentQueued || projection.Assignment.ID != newer.AssignmentID {
			t.Fatalf("recovery leaked durable cancellation: %+v", projection)
		}
	}
}
