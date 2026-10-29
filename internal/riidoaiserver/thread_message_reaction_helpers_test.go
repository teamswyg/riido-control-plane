package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

// seedReactionThread assigns an agent as user-1 and posts a reply as user-2.
// It returns the thread that holds the reply.
func seedReactionThread(t *testing.T) (*DevelopmentAIAgentClientStore, AuthorizationResult, AuthorizationResult, string, string) {
	t.Helper()
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	owner := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	collaborator := AuthorizationResult{PrincipalID: "user-2", WorkspaceID: defaultAIAgentClientWorkspaceID}
	taskID := "task-reaction"
	assigned, err := store.AssignAIAgentTask(ctx, owner, taskID, AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw"})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	reply, err := store.CreateAIAgentTaskThreadMessage(ctx, collaborator, taskID, assigned.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "reply"})
	if err != nil {
		t.Fatalf("CreateAIAgentTaskThreadMessage: %v", err)
	}
	return store, owner, collaborator, taskID, reply.ThreadID
}

func reactionHistoryMessage(t *testing.T, store *DevelopmentAIAgentClientStore, principal AuthorizationResult, taskID, threadID string, role AIAgentTaskThreadMessageRole) AIAgentTaskThreadHistoryMessage {
	t.Helper()
	history, err := store.ListAIAgentTaskThreadHistory(context.Background(), principal, taskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreadHistory: %v", err)
	}
	for _, message := range historyThreadByID(t, history, threadID).Messages {
		if message.Role == role {
			return message
		}
	}
	t.Fatalf("thread %s has no %s message", threadID, role)
	return AIAgentTaskThreadHistoryMessage{}
}

// assertReactions compares groups written as "emoji:kind:id,kind:id".
func assertReactions(t *testing.T, reactions []AIAgentMessageReaction, groups ...string) {
	t.Helper()
	got := make([]string, 0, len(reactions))
	for _, reaction := range reactions {
		actors := make([]string, 0, len(reaction.Actors))
		for _, actor := range reaction.Actors {
			actors = append(actors, string(actor.Kind)+":"+actor.ID)
		}
		if reaction.Count != len(reaction.Actors) {
			t.Fatalf("count %d != actors %d", reaction.Count, len(reaction.Actors))
		}
		got = append(got, reaction.Emoji+":"+strings.Join(actors, ","))
	}
	if strings.Join(got, " ") != strings.Join(groups, " ") {
		t.Fatalf("reactions = %v, want %v", got, groups)
	}
}
