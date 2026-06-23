package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentTaskThreadHistoryRestoresUserMessages(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-history-persist", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	_, err = store.CreateAIAgentTaskThreadMessage(ctx, principal, "task-history-persist", assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		Body:            "persist this follow-up",
		SourceMessageID: "source-message-persist",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen persistent store: %v", err)
	}
	history, err := reopened.ListAIAgentTaskThreadHistory(ctx, principal, "task-history-persist")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory reopened: %v", err)
	}
	thread := historyThreadByID(t, history, assigned.ThreadID)
	if !historyMessagesContainUserBody(thread.Messages, "persist this follow-up") {
		t.Fatalf("reopened history messages = %+v", thread.Messages)
	}
}
