package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHTTPMembershipRemovalCancelsLegacyExecutionWithoutGenerationReceipt(t *testing.T) {
	const taskID, agentID, token = "task-legacy-durable-removal", "agent-public-openclaw", "ai-agent-token"
	ctx := context.Background()
	f := newMembershipSharedDynamoFixture(t)
	clientA, storeA := newMembershipDynamoReplica(t, f)
	record := sampleQueuedAssignmentOperationRecord(time.Now().UTC())
	record.TaskID = taskID
	record.AgentID = agentID
	record.AssignmentID = "asn-000001"
	record.Assignment.ID = record.AssignmentID
	record.Assignment.TaskID = taskID
	record.Assignment.AgentID = agentID
	record.Assignment.RuntimeProvider = "openclaw"
	for i := range record.Events {
		record.Events[i].TaskID = taskID
		record.Events[i].AgentID = agentID
		record.Events[i].AssignmentID = record.AssignmentID
	}
	if err := storeA.operationStore.SaveAssignmentOperation(ctx, record); err != nil {
		t.Fatal(err)
	}
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	if _, err := clientA.AssignAIAgentTask(ctx, p, taskID, AssignAIAgentTaskRequest{AgentID: agentID, AssignmentID: record.AssignmentID, durableState: AssignmentQueued}); err != nil {
		t.Fatal(err)
	}
	clientB, storeB := newMembershipDynamoReplica(t, f)
	handler := NewServer(ServerConfig{AIAgentClient: clientB, Assignment: storeB, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}, Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1")}).Handler()
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	reused := createProviderMultiAssignment(t, handler, base, token, taskID, agentID)
	if reused.AssignmentID != "asn-000001" {
		t.Fatalf("legacy additive reuse changed execution: %s", reused.AssignmentID)
	}
	removed := serveUnassignTaskBoundary(handler, base+"/tasks/"+taskID+"/agent-assignments/"+agentID, token, `{"assignment_id":"asn-000001"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("legacy remove=%d %s", removed.Code, removed.Body.String())
	}
	projection, found, err := storeA.operationStore.(AssignmentProjectionReader).LoadAssignmentProjection(ctx, "asn-000001")
	if err != nil || !found || projection.Assignment.State != AssignmentCancelled {
		t.Fatalf("legacy execution continued after accepted removal: %+v err=%v", projection, err)
	}
}
