package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPAmbiguousPublicationFailurePreservesActuallyCommittedExecution(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-ambiguous-publication")
	first := s.stoppedAssignment(t)
	clientB, storeB := newMembershipDynamoReplica(t, s.f)
	publisher := s.replicaHandler(clientB, storeB)
	s.f.mu.Lock()
	s.f.ambiguousScopedCommit = true
	s.f.mu.Unlock()
	posted := s.post(publisher, "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	s.f.mu.Lock()
	s.f.ambiguousScopedCommit = false
	s.f.mu.Unlock()
	if posted.Code == http.StatusAccepted {
		t.Fatal("injected lost response was accepted")
	}
	members := s.participants(t)
	if len(members) != 1 || members[0].LatestAssignmentID == first.AssignmentID || members[0].LatestAssignmentID == "" {
		t.Fatalf("actual publication was lost: %+v", members)
	}
	projection, found, err := s.store.operationStore.(AssignmentProjectionReader).LoadAssignmentProjection(context.Background(), members[0].LatestAssignmentID)
	if err != nil || !found || projection.Assignment.State != AssignmentQueued {
		t.Fatalf("ambiguous error cancelled committed execution: state=%s err=%v", projection.Assignment.State, err)
	}
}
