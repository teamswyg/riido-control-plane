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
