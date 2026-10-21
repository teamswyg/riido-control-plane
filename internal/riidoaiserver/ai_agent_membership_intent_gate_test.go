package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPMembershipRemovalAcceptsLatestIntentGateAfterDurableExecution(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-intent-gate-after-durable")
	s.stoppedAssignment(t)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	intent, err := s.client.CreateAIAgentTaskAgentAssignment(context.Background(), principal, s.taskID, AssignAIAgentTaskRequest{AgentID: s.agentID, intentGateRequired: true})
	if err != nil || !isIntentGateAssignmentID(intent.AssignmentID) {
		t.Fatalf("intent gate publication=%+v err=%v", intent, err)
	}
	participants := s.participants(t)
	if len(participants) != 1 || participants[0].LatestAssignmentID != intent.AssignmentID {
		t.Fatalf("latest retained intent gate missing: %+v", participants)
	}
	removed := serveUnassignTaskBoundary(s.handler, s.base+"/tasks/"+s.taskID+"/agent-assignments/"+s.agentID, s.token, `{"assignment_id":"`+intent.AssignmentID+`"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("latest intent gate remove=%d %s", removed.Code, removed.Body.String())
	}
}
