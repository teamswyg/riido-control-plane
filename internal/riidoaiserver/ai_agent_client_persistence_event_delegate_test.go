package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestPersistentAIAgentClientStoreEventAndThreadDelegates(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-persistent-event", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	thread, err := store.FindAIAgentTaskThreadByID(ctx, principal.WorkspaceID, assigned.ThreadID)
	if err != nil || thread.ThreadID != assigned.ThreadID {
		t.Fatalf("FindAIAgentTaskThreadByID = %+v, %v", thread, err)
	}
	before := snapshots.saves
	err = store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, TaskEvent{
		TaskID:       assigned.TaskID,
		AssignmentID: assigned.AssignmentID,
		AgentID:      assigned.AgentID,
		Type:         EventAssignmentRunning,
		State:        AssignmentRunning,
		Message:      "delegate event running",
		At:           time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent: %v", err)
	}
	if snapshots.saves <= before {
		t.Fatalf("record assignment event did not save: %d <= %d", snapshots.saves, before)
	}
}
