package riidoaiserver

import (
	"context"
	"testing"
)

// A client reacts with the agent message id it rendered before the run state
// changed. The id still means this thread's run.
func TestThreadMessageReactionAcceptsAgentMessageIDFromEarlierState(t *testing.T) {
	ctx := context.Background()
	store, owner, _, taskID, threadID := seedReactionThread(t)
	rendered := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleAgent)
	setReactionThreadState(store, taskID, threadID, AgentAssignmentStateCompleted)

	if _, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, rendered.MessageID, "1f44d", true); err != nil {
		t.Fatalf("react with earlier agent message id: %v", err)
	}
	completed := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleAgent)
	if completed.MessageID == rendered.MessageID {
		t.Fatal("test needs the agent message id to change with the run state")
	}
	assertReactions(t, completed.Reactions, "1f44d:member:user-1")
	if _, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, "msg_user_missing", "1f44d", true); err == nil {
		t.Fatal("an unknown user message id must still fail")
	}
}

// The 👀 write can reach monolith while its record fails. A run that then
// stops still writes the removal once, and a confirmed removal stops it.
func TestCommentStatusReactionRemovesUnconfirmedWrite(t *testing.T) {
	ctx := context.Background()
	store, _, _, taskID, threadID := seedReactionThread(t)
	if err := store.RecordCommentDeliveryEvent(ctx, defaultAIAgentClientWorkspaceID, threadID, "event"); err != nil {
		t.Fatal(err)
	}
	setReactionThreadState(store, taskID, threadID, AgentAssignmentStateStopped)
	plans := store.CommentStatusReactionPlans(taskID)
	if len(plans) != 1 || plans[0].Emoji != "" {
		t.Fatalf("plans = %+v, want one removal", plans)
	}
	if err := store.MarkCommentStatusReactionApplied(ctx, defaultAIAgentClientWorkspaceID, threadID, ""); err != nil {
		t.Fatal(err)
	}
	if plans := store.CommentStatusReactionPlans(taskID); len(plans) != 0 {
		t.Fatalf("confirmed removal planned again: %+v", plans)
	}
}

// A run started by a monolith comment shows its status reaction there; a
// follow-up thread message of the same run gets none.
func TestCommentStartedRunKeepsStatusReactionOffThreadMessages(t *testing.T) {
	store, owner, _, taskID, threadID := seedReactionThread(t)
	if err := store.RecordCommentDeliveryEvent(context.Background(), defaultAIAgentClientWorkspaceID, threadID, "event"); err != nil {
		t.Fatal(err)
	}
	user := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	if len(user.Reactions) != 0 {
		t.Fatalf("comment-started run reacted on a thread message: %+v", user.Reactions)
	}
}
