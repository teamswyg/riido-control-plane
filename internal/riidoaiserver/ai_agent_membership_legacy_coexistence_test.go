package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHTTPMembershipRemovalCancelsLegacyAndReceiptExecutionsTogether(t *testing.T) {
	s := newMembershipNativeScenario(t, "task-legacy-receipt-coexistence")
	next := createProviderMultiAssignment(t, s.handler, s.base, s.token, s.taskID, s.agentID)
	record := sampleQueuedAssignmentOperationRecord(time.Now().UTC())
	record.TaskID, record.AgentID, record.AssignmentID = s.taskID, s.agentID, "asn-000001"
	record.Assignment.ID, record.Assignment.TaskID, record.Assignment.AgentID = record.AssignmentID, s.taskID, s.agentID
	record.Assignment.RuntimeProvider = "openclaw"
	for i := range record.Events {
		record.Events[i].TaskID, record.Events[i].AgentID, record.Events[i].AssignmentID = s.taskID, s.agentID, record.AssignmentID
	}
	if err := s.store.operationStore.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	client, store := newMembershipDynamoReplica(t, s.f)
	removed := serveUnassignTaskBoundary(s.replicaHandler(client, store), s.base+"/tasks/"+s.taskID+"/agent-assignments/"+s.agentID, s.token, `{"assignment_id":"`+next.AssignmentID+`"}`)
	if removed.Code != http.StatusAccepted {
		t.Fatalf("coexisting remove=%d %s", removed.Code, removed.Body.String())
	}
	for _, id := range []string{record.AssignmentID, next.AssignmentID} {
		projection, found, err := s.store.LoadAssignmentProjection(context.Background(), id)
		if err != nil || !found || projection.Assignment.State != AssignmentCancelled {
			t.Fatalf("execution %s continued after accepted removal: %+v err=%v", id, projection, err)
		}
	}
}
