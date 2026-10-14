package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPMembershipRemovalRejectsStaleIntentGateAfterNewDurableExecution(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-stale-intent-gate")
	s.stoppedAssignment(t)
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	intent, err := s.client.CreateAIAgentTaskAgentAssignment(context.Background(), p, s.taskID, AssignAIAgentTaskRequest{AgentID: s.agentID, intentGateRequired: true})
	if err != nil {
		t.Fatal(err)
	}
	stale, _ := newMembershipDynamoReplica(t, s.f)
	if err := stale.reloadSnapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	posted := s.post(s.handler, "/assignment", `{"agent_id":"`+s.agentID+`"}`)
	if posted.Code != http.StatusAccepted {
		t.Fatalf("new execution=%d %s", posted.Code, posted.Body.String())
	}
	var next AIAgentTaskActionResponse
	if err := json.Unmarshal(posted.Body.Bytes(), &next); err != nil {
		t.Fatal(err)
	}
	stale.mu.Lock()
	stale.taskThreads[s.taskID][len(stale.taskThreads[s.taskID])-1].Message = "Cached progress update"
	stale.mu.Unlock()
	if err := stale.saveSnapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	participants := s.participants(t)
	if len(participants) != 1 || participants[0].LatestAssignmentID == intent.AssignmentID {
		t.Fatalf("stale intent ID exposed as latest: %+v", participants)
	}
	removed := serveUnassignTaskBoundary(s.handler, s.base+"/tasks/"+s.taskID+"/agent-assignments/"+s.agentID, s.token, `{"assignment_id":"`+intent.AssignmentID+`"}`)
	projection, found, err := s.store.LoadAssignmentProjection(context.Background(), next.AssignmentID)
	if err != nil || !found || projection.Assignment.State != AssignmentQueued {
		t.Fatalf("new execution cancelled by stale intent ID: %+v err=%v", projection, err)
	}
	if removed.Code != http.StatusConflict {
		t.Fatalf("stale intent remove=%d %s", removed.Code, removed.Body.String())
	}
}
