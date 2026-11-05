package riidoaiserver

import (
	"errors"
	"testing"
)

func TestMessageReactionsPersistAndSeparateMembersFromAgentStatus(t *testing.T) {
	ctx := t.Context()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	root, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "reaction-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "reaction-run", conversationID: "conversation", Instruction: "Continue", SourceCommentID: "source", SourceMessageID: "source"})
	if err != nil {
		t.Fatal(err)
	}
	reply := root
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, root.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	for _, row := range history.Threads {
		for _, message := range row.Messages {
			if message.SourceMessageID == "source" {
				id = message.MessageID
				if len(message.Reactions) != 1 || message.Reactions[0].Emoji != "1f440" {
					t.Fatalf("accepted reaction: %+v", message)
				}
			}
		}
	}
	if id == "" {
		t.Fatal("missing message")
	}
	for range 2 {
		response, err := store.MutateTaskMessageReaction(ctx, principal, root.TaskID, reply.ThreadID, id, "1F440", false)
		if err != nil {
			t.Fatal(err)
		}
		if len(response.Reactions) != 1 || response.Reactions[0].Count != 2 {
			t.Fatalf("member+agent merge: %+v", response)
		}
	}
	reopened := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	response, err := reopened.MutateTaskMessageReaction(ctx, principal, root.TaskID, reply.ThreadID, id, "1f440", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Reactions) != 1 || response.Reactions[0].Count != 1 || response.Reactions[0].Actors[0].Kind != "agent" {
		t.Fatalf("delete affected agent: %+v", response)
	}
	if _, err := reopened.MutateTaskMessageReaction(ctx, principal, "wrong-task", reply.ThreadID, id, "2705", false); !errors.Is(err, ErrAIAgentNotFound) {
		t.Fatalf("cross task: %v", err)
	}
	outsider := AuthorizationResult{PrincipalID: "outsider", WorkspaceID: "wrong-workspace"}
	if _, err := reopened.MutateTaskMessageReaction(ctx, outsider, root.TaskID, reply.ThreadID, id, "2705", false); !errors.Is(err, ErrAIAgentNotFound) {
		t.Fatalf("cross workspace: %v", err)
	}
	for _, state := range []struct {
		work  AgentWorkStatus
		state AgentAssignmentState
		emoji string
	}{{AgentWorkStatusCompleted, "completed", "2705"}, {AgentWorkStatusFailed, "failed", "274c"}, {AgentWorkStatusIdle, "cancelled", ""}} {
		reopened.mu.Lock()
		rows := reopened.taskThreads[root.TaskID]
		for i := range rows {
			if rows[i].ThreadID == reply.ThreadID {
				rows[i].WorkStatus = state.work
				rows[i].AssignmentState = state.state
			}
		}
		reopened.taskThreads[root.TaskID] = rows
		reopened.mu.Unlock()
		history, err := reopened.ListAIAgentTaskThreadHistory(ctx, principal, root.TaskID)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range history.Threads {
			for _, message := range row.Messages {
				if message.MessageID == id {
					if state.emoji == "" {
						if len(message.Reactions) != 0 {
							t.Fatal("cancel retained status")
						}
					} else if len(message.Reactions) != 1 || message.Reactions[0].Emoji != state.emoji {
						t.Fatalf("status %s: %+v", state.emoji, message.Reactions)
					}
				}
			}
		}
	}
}

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

func TestMessageReactionKeepsOriginalTriggerDuringRunningFollowup(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "reaction-binding", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", conversationID: "conversation", SourceCommentID: "source-a", SourceMessageID: "source-a", Instruction: "A"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateAIAgentTaskThreadMessage(t.Context(), principal, action.TaskID, action.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "B", SourceMessageID: "source-b"})
	if err != nil {
		t.Fatal(err)
	}
	for _, status := range []struct {
		work  AgentWorkStatus
		state AgentAssignmentState
		emoji string
	}{{AgentWorkStatusRunning, "running", "1f440"}, {AgentWorkStatusCompleted, "completed", "2705"}} {
		store.mu.Lock()
		store.taskThreads[action.TaskID][0].WorkStatus = status.work
		store.taskThreads[action.TaskID][0].AssignmentState = status.state
		store.mu.Unlock()
		history, err := store.ListAIAgentTaskThreadHistory(t.Context(), principal, action.TaskID)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range history.Threads {
			for _, m := range row.Messages {
				if m.SourceMessageID == "source-a" && (len(m.Reactions) != 1 || m.Reactions[0].Emoji != status.emoji) {
					t.Fatalf("original trigger lost reaction: %+v", m)
				}
				if m.SourceMessageID == "source-b" && len(m.Reactions) != 0 {
					t.Fatalf("undelivered followup got reaction: %+v", m)
				}
			}
		}
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
