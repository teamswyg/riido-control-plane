package riidoaiserver

import (
	"context"
	"testing"
)

// A progress batch without run_id rewrites the thread run id. Reactions follow
// the assignment, so the run keeps its status reaction and member reactions.
func TestThreadMessageReactionSurvivesProgressRunIDRewrite(t *testing.T) {
	ctx := context.Background()
	store, owner, _, taskID, threadID := seedReactionThread(t)
	agent := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleAgent)
	if _, err := store.ReactToAIAgentTaskThreadMessage(ctx, owner, taskID, threadID, agent.MessageID, "1f44d", true); err != nil {
		t.Fatalf("react: %v", err)
	}
	thread, ok := store.rawTaskThreadByIDLocked(taskID, threadID)
	if !ok {
		t.Fatal("thread missing")
	}
	if _, err := store.RecordAIAgentThreadProgress(ctx, thread.AgentID, AgentThreadProgressBatchRequest{
		TaskID: taskID, AssignmentID: thread.AssignmentID, Lines: []AgentThreadProgressLine{{Seq: 1, Message: "working"}},
	}); err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	if rewritten, _ := store.rawTaskThreadByIDLocked(taskID, threadID); rewritten.RunID == thread.RunID {
		t.Fatalf("test needs the batch to rewrite run id %q", thread.RunID)
	}

	user := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleUser)
	assertReactions(t, user.Reactions, "1f440:agent:agent-public-openclaw")
	running := reactionHistoryMessage(t, store, owner, taskID, threadID, AIAgentTaskThreadMessageRoleAgent)
	assertReactions(t, running.Reactions, "1f44d:member:user-1")
}
