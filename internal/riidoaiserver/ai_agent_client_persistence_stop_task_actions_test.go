package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentClientStorePersistsTaskStopByAssignment(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	first, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-persist-task-stop", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-persist-task-stop-1",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment first: %v", err)
	}
	second, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, first.TaskID, AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-persist-task-stop-2",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment second: %v", err)
	}
	stopped, err := store.StopAIAgentTask(ctx, principal, first.TaskID, StopAIAgentTaskRequest{
		AssignmentID: first.AssignmentID,
		Reason:       "target stop",
	})
	if err != nil {
		t.Fatalf("StopAIAgentTask: %v", err)
	}
	if stopped.AssignmentID != first.AssignmentID ||
		stopped.AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("stop response = %+v", stopped)
	}

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen persistent store: %v", err)
	}
	threads, err := reopened.ListAIAgentTaskThreads(ctx, principal, first.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if actionThreadByID(t, threads, first.ThreadID).AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("first assignment was not restored as stopped: %+v", threads.Threads)
	}
	if actionThreadByID(t, threads, second.ThreadID).AssignmentState == AgentAssignmentStateStopped {
		t.Fatalf("sibling assignment was stopped unexpectedly: %+v", threads.Threads)
	}
}
