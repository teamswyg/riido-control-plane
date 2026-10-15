package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPAmbiguousAcceptedPublicationSurvivesNewerGeneration(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-accepted-publication-newer-generation")
	s.stoppedAssignment(t)
	clientB, storeB := newMembershipDynamoReplica(t, s.f)
	clientC, storeC := newMembershipDynamoReplica(t, s.f)
	entered, resume := make(chan struct{}), make(chan struct{})
	s.f.mu.Lock()
	s.f.ambiguousScopedCommit = true
	s.f.beforeGenerationRead = func() {
		s.f.mu.Lock()
		s.f.beforeGenerationRead = nil
		s.f.ambiguousScopedCommit = false
		s.f.mu.Unlock()
		close(entered)
		<-resume
	}
	s.f.mu.Unlock()
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		result <- s.post(s.replicaHandler(clientB, storeB), "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		close(resume)
		t.Fatal("accepted publication did not reach recovery observation")
	}
	newer := createProviderMultiAssignment(t, s.replicaHandler(clientC, storeC), s.base, s.token, s.taskID, s.agentID)
	close(resume)
	if (<-result).Code == http.StatusAccepted {
		t.Fatal("lost publication response unexpectedly succeeded")
	}
	members := s.participants(t)
	if len(members) != 1 || members[0].LatestAssignmentID != newer.AssignmentID {
		t.Fatalf("newer membership changed: %+v", members)
	}
	_, active, err := s.store.loadTaskAgentDurableExecutions(context.Background(), s.taskID, s.agentID)
	if err != nil || len(active) != 2 {
		t.Fatalf("accepted older execution cancelled: active=%+v err=%v", active, err)
	}
	for _, projection := range active {
		if projection.Assignment.State != AssignmentQueued {
			t.Fatalf("accepted execution changed: %+v", projection)
		}
	}
}
