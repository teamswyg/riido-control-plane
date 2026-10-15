package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPPublicationRetriesWhenExecutionFinishesBeforeReceiptCommit(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-publication-terminal-receipt-race")
	first := s.stoppedAssignment(t)
	s.f.mu.Lock()
	s.f.beforeScopedCommit = func(snapshot AIAgentClientSnapshot) {
		s.f.mu.Lock()
		s.f.beforeScopedCommit = nil
		s.f.mu.Unlock()
		assignmentID, _, _ := strings.Cut(snapshot.TaskAgentMemberships[s.taskID][s.agentID].DurableGeneration, "@")
		projection, found, err := s.store.LoadAssignmentProjection(context.Background(), assignmentID)
		if err != nil || !found {
			t.Errorf("missing publication execution: %v", err)
			return
		}
		advanceFailedPublicationExecution(t, s.store, projection.Assignment, "completed")
	}
	s.f.mu.Unlock()
	posted := s.post(s.handler, "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	if posted.Code != http.StatusAccepted {
		t.Fatalf("terminal receipt race rejected publication: %d %s", posted.Code, posted.Body.String())
	}
	members := s.participants(t)
	if len(members) != 1 || members[0].LatestAssignmentID == first.AssignmentID {
		t.Fatalf("completed participant missing: %+v", members)
	}
	projection, found, err := s.store.LoadAssignmentProjection(context.Background(), members[0].LatestAssignmentID)
	if err != nil || !found || projection.Assignment.State != AssignmentCompleted {
		t.Fatalf("completed execution changed: %+v err=%v", projection, err)
	}
	_, active, err := s.store.loadTaskAgentDurableExecutions(context.Background(), s.taskID, s.agentID)
	if err != nil || len(active) != 0 {
		t.Fatalf("terminal receipt recreated: %+v err=%v", active, err)
	}
}
