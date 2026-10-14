package riidoaiserver

import (
	"context"
	"testing"
)

func TestAIAgentTaskThreadHistoryPreservesReplyAuthor(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{
		PrincipalID: "user-1",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	collaborator := AuthorizationResult{
		PrincipalID: "user-2",
		WorkspaceID: defaultAIAgentClientWorkspaceID,
	}
	assigned, err := store.AssignAIAgentTask(ctx, owner, "task-history-author", AssignAIAgentTaskRequest{
		AgentID: "agent-public-openclaw",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	body := "다른 사용자가 남긴 답글"
	_, err = store.CreateAIAgentTaskThreadMessage(
		ctx,
		collaborator,
		"task-history-author",
		assigned.ThreadID,
		CreateAIAgentTaskThreadMessageRequest{Body: body},
	)
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	history, err := store.ListAIAgentTaskThreadHistory(ctx, owner, "task-history-author")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	thread := historyThreadByID(t, history, assigned.ThreadID)
	if !historyMessagesContainUserAuthor(thread.Messages, body, collaborator.PrincipalID) {
		t.Fatalf("history messages missing collaborator author: %+v", thread.Messages)
	}
}

func historyMessagesContainUserAuthor(
	messages []AIAgentTaskThreadHistoryMessage,
	body string,
	authorPrincipalID string,
) bool {
	for _, message := range messages {
		if message.Role == AIAgentTaskThreadMessageRoleUser &&
			message.Body == body &&
			message.AuthorPrincipalID == authorPrincipalID {
			return true
		}
	}
	return false
}
