package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPRejectedCreationDoesNotBecomeReconciledSuccess(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-rejected-creation")
	s.f.mu.Lock()
	s.f.rejectCreation = true
	s.f.mu.Unlock()
	posted := s.post(s.handler, "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	if posted.Code == http.StatusAccepted {
		t.Fatal("uncommitted creation reported as success")
	}
	generation, active, err := s.store.loadTaskAgentDurableExecutions(context.Background(), s.taskID, s.agentID)
	if err != nil || generation != "" || len(active) != 0 || len(s.participants(t)) != 0 {
		t.Fatalf("uncommitted creation became visible: generation=%s active=%v err=%v", generation, active, err)
	}
}
