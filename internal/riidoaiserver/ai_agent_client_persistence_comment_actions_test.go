package riidoaiserver

import (
	"context"
	"testing"
)

func TestPersistentAIAgentClientStorePersistsTaskCommentUserMessage(t *testing.T) {
	ctx := context.Background()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("OpenPersistentAIAgentClientStore: %v", err)
	}
	assigned, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "task-persist-comment", AssignAIAgentTaskRequest{
		AgentID: "agent-public-openclaw",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskAgentAssignment: %v", err)
	}
	const body = "persist queued follow-up"
	if _, err := store.SubmitAIAgentTaskComment(ctx, principal, assigned.TaskID, SubmitAIAgentTaskCommentRequest{
		AgentID: assigned.AgentID,
		Body:    body,
	}); err != nil {
		t.Fatalf("SubmitAIAgentTaskComment: %v", err)
	}

	reopened, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), snapshots)
	if err != nil {
		t.Fatalf("reopen persistent store: %v", err)
	}
	history, err := reopened.ListAIAgentTaskThreadHistory(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	if !historyContainsUserBody(history, body) {
		t.Fatalf("persisted history missing %q: %+v", body, history.Threads)
	}
}

func historyContainsUserBody(history AIAgentTaskThreadHistoryCollectionResponse, body string) bool {
	for _, thread := range history.Threads {
		if historyMessagesContainUserBody(thread.Messages, body) {
			return true
		}
	}
	return false
}
