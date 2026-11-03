package riidoaiserver

import (
	"testing"
)

// RIID-5909: a run started by another agent's reply reacts on that reply in
// the caller's thread, not on its own thread.
func TestAgentInvokedRunReactsOnCallerReply(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	source := "reply-a"
	store.taskThreads["task"] = []AIAgentTaskThreadRecord{{
		TaskID: "task", ThreadID: "thread-b", AgentID: "agent-b", AssignmentID: "run-b",
		SourceMessageID: source, ReactionSourceMessageID: &source,
		CommentInvocation: &CommentInvocation{OriginThreadID: "thread-a", OriginMessageID: source},
	}}
	for _, tc := range []struct {
		work AgentWorkStatus
		want string
	}{{AgentWorkStatusRunning, "1f440"}, {AgentWorkStatusCompleted, "2705"}, {AgentWorkStatusFailed, "274c"}} {
		store.taskThreads["task"][0].WorkStatus = tc.work
		records := []AIAgentTaskThreadHistoryRecord{
			{TaskID: "task", ThreadID: "thread-a", Messages: []AIAgentTaskThreadHistoryMessage{{MessageID: source, Role: AIAgentTaskThreadMessageRoleAgent}}},
			{TaskID: "task", ThreadID: "thread-b", Messages: []AIAgentTaskThreadHistoryMessage{{MessageID: "answer-b", Role: AIAgentTaskThreadMessageRoleAgent}}},
		}
		store.projectAgentMessageReactionsLocked(records)
		got := records[0].Messages[0].Reactions
		if len(got) != 1 || got[0].Emoji != tc.want || len(got[0].Actors) != 1 || got[0].Actors[0] != (AIAgentMessageReactionActor{Kind: "agent", ID: "agent-b"}) {
			t.Fatalf("%s caller reply reactions = %+v, want only %s by agent-b", tc.work, got, tc.want)
		}
		if own := records[1].Messages[0].Reactions; len(own) != 0 {
			t.Fatalf("%s own thread reactions = %+v", tc.work, own)
		}
	}
}
