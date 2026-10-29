package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

func TestThreadMessageReactionsMemberAndAgentStatus(t *testing.T) {
	ctx := context.Background()
	store, owner, collaborator, taskID, threadID := seedReactionThread(t)
	user := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	assertReactions(t, user.Reactions, "1f440:agent:agent-public-openclaw")

	first, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, user.MessageID, "1f440", true)
	if err != nil || !first.Changed || first.Notify == nil || first.Notify.RecipientPrincipalID != collaborator.PrincipalID || first.Notify.Emoji != "1f440" {
		t.Fatalf("first add = %+v, %v", first, err)
	}
	again, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, user.MessageID, "1f440", true)
	if err != nil || again.Changed || again.Notify != nil {
		t.Fatalf("repeated add must be a no-op: %+v, %v", again, err)
	}
	assertReactions(t, again.Response.Reactions, "1f440:member:user-1,agent:agent-public-openclaw")

	own, err := store.ReactToAIAgentTaskThreadMessage(ctx, collaborator, taskID, threadID, user.MessageID, "1f44d", true)
	if err != nil || own.Notify != nil {
		t.Fatalf("own message must not notify: %+v, %v", own, err)
	}
	deleted, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, user.MessageID, "1f440", false)
	if err != nil || !deleted.Changed || deleted.Notify == nil || deleted.Notify.Emoji != "" {
		t.Fatalf("delete = %+v, %v", deleted, err)
	}
	assertReactions(t, deleted.Response.Reactions, "1f44d:member:user-2", "1f440:agent:agent-public-openclaw")
	missing, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, user.MessageID, "1f440", false)
	if err != nil || missing.Changed {
		t.Fatalf("deleting a missing reaction must succeed without change: %+v, %v", missing, err)
	}
}

func TestThreadMessageReactionsRejectInvalidTargets(t *testing.T) {
	ctx := context.Background()
	store, owner, _, taskID, threadID := seedReactionThread(t)
	user := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	if _, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, user.MessageID, "1F440", true); !errors.Is(err, errThreadMessageReactionEmoji) {
		t.Fatalf("uppercase emoji error = %v", err)
	}
	if _, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, "msg_missing", "1f440", true); !errors.Is(err, ErrAIAgentNotFound) {
		t.Fatalf("missing message error = %v", err)
	}
	progress := AIAgentTaskThreadHistoryMessage{MessageID: "msg_progress_x", Role: AIAgentTaskThreadMessageRoleProgress}
	if _, ok := taskThreadReactionTarget(progress); ok {
		t.Fatal("progress messages must not be reactable")
	}
}

func TestAgentStatusReactionEmojiFollowsRunState(t *testing.T) {
	for state, want := range map[AgentAssignmentState]string{
		AgentAssignmentStateQueued: "1f440", AgentAssignmentStateRunning: "1f440", AgentAssignmentStateWaiting: "1f440",
		AgentAssignmentStateCompleted: "2705", AgentAssignmentStateFailed: "274c",
		AgentAssignmentStateStopping: "", AgentAssignmentStateStopped: "", AgentAssignmentStateUnassigned: "",
	} {
		if got := agentStatusReactionEmoji(state); got != want {
			t.Fatalf("%s = %q, want %q", state, got, want)
		}
	}
}
