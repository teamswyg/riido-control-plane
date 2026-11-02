package riidoaiserver

import (
	"testing"
)

func TestMessageReactionRejectsProgressAndInvalidEmoji(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root, err := store.AssignAIAgentTask(t.Context(), principal, "reaction-progress", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex"})
	if err != nil {
		t.Fatal(err)
	}
	store.taskThreadMessages[root.ThreadID] = []AIAgentTaskThreadHistoryMessage{{MessageID: "progress", Role: AIAgentTaskThreadMessageRoleProgress, Body: "thinking"}}
	if _, err := store.MutateTaskMessageReaction(t.Context(), principal, root.TaskID, root.ThreadID, "progress", "2705", false); err == nil {
		t.Fatal("progress accepted")
	}
	if _, err := normalizeMessageReactionEmoji("✅"); err == nil {
		t.Fatal("invalid format accepted")
	}
}

func TestMessageReactionEmojiPresentationIdentity(t *testing.T) {
	for _, input := range []string{"2764", "2764-fe0f", "2764-FE0F"} {
		got, err := normalizeMessageReactionEmoji(input)
		if err != nil || got != "2764-fe0f" {
			t.Fatalf("%q: %q %v", input, got, err)
		}
	}
}
