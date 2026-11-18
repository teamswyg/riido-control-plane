package riidoaiserver

import "testing"

func TestCompletedResultRepairPersistsAcrossIndependentReaders(t *testing.T) {
	f, first, _, _ := terminalResultFixture(t)
	seedMissingTerminalResult(t, f, first)
	dynamo := newMembershipSharedDynamoFixture(t)
	writer, _ := newMembershipDynamoReplica(t, dynamo)
	writer.taskThreads[first.TaskID] = []AIAgentTaskThreadRecord{copyTaskThread(f.client.taskThreads[first.TaskID][0])}
	if err := writer.saveSnapshot(t.Context()); err != nil {
		t.Fatal(err)
	}
	reader, _ := newMembershipDynamoReplica(t, dynamo)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	changed, err := reader.ReconcileAIAgentActiveThreadProjections(t.Context(), principal, first.TaskID, f.store)
	if err != nil || !changed {
		t.Fatal("persistent repair failed", err)
	}
	reloaded, _ := newMembershipDynamoReplica(t, dynamo)
	if reloaded.taskThreads[first.TaskID][0].ResultMessage != terminalResultBody {
		t.Fatal("a fresh server lost the repaired result")
	}
	changed, err = reloaded.ReconcileAIAgentActiveThreadProjections(t.Context(), principal, first.TaskID, f.store)
	if err != nil || changed {
		t.Fatal("repeated reads must be idempotent", err)
	}
}
