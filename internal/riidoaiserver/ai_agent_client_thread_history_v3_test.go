package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentTaskThreadHistoryKeepsUserFollowupMessages(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-history", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	followup, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, "task-history", assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{
		Body:            `<p>다시 해줘</p>`,
		SourceMessageID: "source-message-1",
	})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, "task-history")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	thread := historyThreadByID(t, history, followup.ThreadID)
	if thread.AgentSnapshotID == "" || len(history.AgentSnapshots) != 1 {
		t.Fatalf("agent snapshot map = id %q map %+v", thread.AgentSnapshotID, history.AgentSnapshots)
	}
	if !historyMessagesContainUserBody(thread.Messages, `<p>다시 해줘</p>`) {
		t.Fatalf("history messages missing user body: %+v", thread.Messages)
	}
}

func historyThreadByID(t *testing.T, history AIAgentTaskThreadHistoryCollectionResponse, threadID string) AIAgentTaskThreadHistoryRecord {
	t.Helper()
	for _, thread := range history.Threads {
		if thread.ThreadID == threadID {
			return thread
		}
	}
	t.Fatalf("thread %q not found in %+v", threadID, history.Threads)
	return AIAgentTaskThreadHistoryRecord{}
}

func historyMessagesContainUserBody(messages []AIAgentTaskThreadHistoryMessage, body string) bool {
	for _, message := range messages {
		if message.Role == AIAgentTaskThreadMessageRoleUser && message.Body == body {
			return true
		}
	}
	return false
}
