package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecoveryOnlyProjectionDoesNotLoadOrSaveSnapshot(t *testing.T) {
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	loads, saves := snapshots.loads, snapshots.saves
	for range 25 {
		if err := store.RecordAIAgentAssignmentEvent(context.Background(), "agent-owned-codex", AgentEventRequest{}, recoveryProjectionEvent()); err != nil {
			t.Fatal(err)
		}
	}
	if snapshots.loads != loads || snapshots.saves != saves {
		t.Fatalf("recovery-only projection performed unnecessary snapshot I/O: loads=%d saves=%d", snapshots.loads-loads, snapshots.saves-saves)
	}
}

func TestRecoveryProjectionPreservesValidationAndCancellation(t *testing.T) {
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.RecordAIAgentAssignmentEvent(ctx, "agent-owned-codex", AgentEventRequest{}, recoveryProjectionEvent()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled recovery = %v", err)
	}
	for _, missing := range []string{"agent", "task", "assignment"} {
		event := recoveryProjectionEvent()
		agentID := "agent-owned-codex"
		switch missing {
		case "agent":
			agentID = ""
		case "task":
			event.TaskID = ""
		case "assignment":
			event.AssignmentID = ""
		}
		if err := store.RecordAIAgentAssignmentEvent(context.Background(), agentID, AgentEventRequest{}, event); err == nil {
			t.Fatalf("missing %s accepted", missing)
		}
	}
}

func TestRecoveryInRequestCannotSkipOrdinaryProjection(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "ordinary-recovery-boundary", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatal(err)
	}
	before := snapshots.saves
	event := TaskEvent{TaskID: assigned.TaskID, AssignmentID: assigned.AssignmentID, AgentID: assigned.AgentID, State: AssignmentRunning, Type: EventAssignmentRunning, Message: "ordinary state", At: time.Now().UTC()}
	req := AgentEventRequest{Metadata: recoveryProjectionEvent().Metadata}
	if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, req, event); err != nil {
		t.Fatal(err)
	}
	if snapshots.saves <= before {
		t.Fatal("uncommitted request metadata skipped ordinary projection persistence")
	}
}
