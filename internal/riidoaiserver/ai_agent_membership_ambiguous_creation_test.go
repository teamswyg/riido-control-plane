package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPAmbiguousCreationResponseReconcilesBeforePublication(t *testing.T) {
	for _, response := range []string{"server-error", "connection-lost"} {
		for _, mode := range []string{"replacement", "additive", "reply"} {
			t.Run(response+"/"+mode, func(t *testing.T) {
				s := newMembershipNativeScenario(t, "task-ambiguous-creation")
				first := s.stoppedAssignment(t)
				suffix, body := "/assignment", `{"agent_id":"`+s.agentID+`"}`
				if mode == "additive" {
					suffix = "/agent-assignments"
				}
				if mode == "reply" {
					suffix, body = "/threads/"+first.ThreadID+"/messages", `{"body":"continue this task"}`
				}
				s.f.mu.Lock()
				s.f.ambiguousCreationCommit = response == "server-error"
				s.f.loseCreationResponse = response == "connection-lost"
				s.f.mu.Unlock()
				posted := s.post(s.handler, suffix, body)
				s.f.mu.Lock()
				s.f.ambiguousCreationCommit = false
				s.f.loseCreationResponse = false
				s.f.mu.Unlock()
				if posted.Code != http.StatusAccepted {
					t.Fatalf("committed creation was reported failed: %d %s", posted.Code, posted.Body.String())
				}
				members := s.participants(t)
				if len(members) != 1 || members[0].LatestAssignmentID == "" || members[0].LatestAssignmentID == first.AssignmentID {
					t.Fatalf("new creation was not published: %+v", members)
				}
				removed := serveUnassignTaskBoundary(s.handler, s.base+"/tasks/"+s.taskID+"/agent-assignments/"+s.agentID, s.token, `{"assignment_id":"`+members[0].LatestAssignmentID+`"}`)
				if removed.Code != http.StatusAccepted {
					t.Fatalf("remove=%d %s", removed.Code, removed.Body.String())
				}
				_, active, err := s.store.loadTaskAgentDurableExecutions(context.Background(), s.taskID, s.agentID)
				if err != nil || len(active) != 0 || len(s.participants(t)) != 0 {
					t.Fatalf("removal left active executions: %v err=%v", active, err)
				}
			})
		}
	}
}
