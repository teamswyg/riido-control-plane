package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHTTPMembershipRemovalDoesNotCancelInvisiblePrivateAgent(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-private-member-removal")
	ctx := context.Background()
	s.client.mu.Lock()
	agent := s.client.agents[s.agentID]
	agent.Visibility = AgentVisibilityPrivate
	agent.OwnerPrincipalID = "user-2"
	s.client.agents[s.agentID] = agent
	s.client.mu.Unlock()
	if err := s.client.saveSnapshot(ctx); err != nil {
		t.Fatal(err)
	}
	record := sampleQueuedAssignmentOperationRecord(time.Now().UTC())
	record.TaskID, record.AgentID = s.taskID, s.agentID
	record.AssignmentID = "asn-000001"
	record.Assignment.ID = record.AssignmentID
	record.Assignment.TaskID = s.taskID
	record.Assignment.AgentID = s.agentID
	record.Assignment.RuntimeProvider = "openclaw"
	for i := range record.Events {
		record.Events[i].TaskID = s.taskID
		record.Events[i].AgentID = s.agentID
		record.Events[i].AssignmentID = record.AssignmentID
	}
	if err := s.store.operationStore.SaveAssignmentOperation(ctx, record); err != nil {
		t.Fatal(err)
	}
	p := AuthorizationResult{PrincipalID: "user-2", WorkspaceID: defaultAIAgentClientWorkspaceID}
	if _, err := s.client.AssignAIAgentTask(ctx, p, s.taskID, AssignAIAgentTaskRequest{AgentID: s.agentID, AssignmentID: record.AssignmentID, durableState: AssignmentQueued}); err != nil {
		t.Fatal(err)
	}
	clientB, storeB := newMembershipDynamoReplica(t, s.f)
	remover := s.replicaHandler(clientB, storeB)
	removed := serveUnassignTaskBoundary(remover, s.base+"/tasks/"+s.taskID+"/agent-assignments/"+s.agentID, s.token, `{}`)
	projection, found, err := s.store.operationStore.(AssignmentProjectionReader).LoadAssignmentProjection(ctx, "asn-000001")
	if err != nil || !found || projection.Assignment.State != AssignmentQueued {
		t.Fatalf("invisible private execution was cancelled: state=%s err=%v", projection.Assignment.State, err)
	}
	if removed.Code != http.StatusNotFound {
		t.Fatalf("private remove=%d %s", removed.Code, removed.Body.String())
	}
}
