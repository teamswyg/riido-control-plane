package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentTaskThreadHistoryKeepsRepeatedUserBodies(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-history-repeat", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	for range 2 {
		_, err := store.CreateAIAgentTaskThreadMessage(ctx, principal, "task-history-repeat", assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{
			Body: "같은 요청",
		})
		if err != nil {
			t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
		}
	}
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, "task-history-repeat")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	thread := historyThreadByID(t, history, assigned.ThreadID)
	if got := countUserHistoryMessages(thread.Messages, "같은 요청"); got != 2 {
		t.Fatalf("repeated user messages = %d, want 2: %+v", got, thread.Messages)
	}
}

func countUserHistoryMessages(messages []AIAgentTaskThreadHistoryMessage, body string) int {
	count := 0
	for _, message := range messages {
		if message.Role == AIAgentTaskThreadMessageRoleUser && message.Body == body {
			count++
		}
	}
	return count
}
