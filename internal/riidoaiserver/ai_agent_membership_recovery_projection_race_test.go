package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestHTTPFailedPublicationRecoveryPreservesConcurrentTerminalResult(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-recovery-terminal-progress")
	first := s.stoppedAssignment(t)
	clientB, storeB := newMembershipDynamoReplica(t, s.f)
	_, daemon := newMembershipDynamoReplica(t, s.f)
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
	go func() {
		result <- s.post(s.replicaHandler(clientB, storeB), "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		close(resume)
		t.Fatal("recovery did not reach cancellation commit")
	}
	_, active, err := daemon.loadTaskAgentDurableExecutions(context.Background(), s.taskID, s.agentID)
	if err != nil || len(active) != 1 {
		t.Fatalf("pending execution=%+v err=%v", active, err)
	}
	advanceFailedPublicationExecution(t, daemon, active[0].Assignment, "completed")
	before, _, err := daemon.LoadAssignmentProjection(context.Background(), active[0].Assignment.ID)
	if err != nil {
		t.Fatal(err)
	}
	close(resume)
	if (<-result).Code == http.StatusAccepted {
		t.Fatal("failed publication was accepted")
	}
	after, _, err := daemon.LoadAssignmentProjection(context.Background(), active[0].Assignment.ID)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("recovery changed terminal result: before=%+v after=%+v err=%v", before, after, err)
	}
	removed := serveUnassignTaskBoundary(s.handler, s.base+"/tasks/"+s.taskID+"/agent-assignments/"+s.agentID, s.token, `{"assignment_id":"`+first.AssignmentID+`"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("terminal progress left failed publication blocked: remove=%d %s", removed.Code, removed.Body.String())
	}
}
