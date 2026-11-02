package riidoaiserver

import (
	"testing"
)

// The started push stays once per assignment after the event buffer no longer
// holds the first notification, including across a restart.
func TestStartedNotificationStaysOnceAfterEventsRollOff(t *testing.T) {
	ctx := t.Context()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "started-once", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", conversationID: "conversation", Instruction: "Continue", SourceCommentID: "source", SourceMessageID: "source"})
	if err != nil {
		t.Fatal(err)
	}
	if got := startedNotifications(store.DevelopmentAIAgentClientStore, action.AssignmentID); got != 1 {
		t.Fatalf("first started notifications = %d, want 1", got)
	}

	reopened := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	reopened.mu.Lock()
	reopened.events = nil
	reopened.appendTaskNotificationLocked(TaskEvent{EventID: "resumed", TaskID: action.TaskID, AssignmentID: action.AssignmentID, WorkspaceID: principal.WorkspaceID, CreatedBy: principal.PrincipalID, State: AssignmentRunning}, AIAgentTaskActionResponse{WorkStatus: AgentWorkStatusRunning})
	reopened.mu.Unlock()
	if got := startedNotifications(reopened.DevelopmentAIAgentClientStore, action.AssignmentID); got != 0 {
		t.Fatalf("started sent again after the events rolled off: %d", got)
	}
}
