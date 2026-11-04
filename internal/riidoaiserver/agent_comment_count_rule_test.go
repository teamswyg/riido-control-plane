package riidoaiserver

import "testing"

// RIID-5842: the list count mirrors what the comments tab shows.
func TestCountVisibleAgentCommentMessagesMirrorsCommentsTab(t *testing.T) {
	threads := []AIAgentTaskThreadHistoryRecord{
		{ThreadID: "comment-run", Messages: []AIAgentTaskThreadHistoryMessage{
			{MessageID: "copied-comment", Role: AIAgentTaskThreadMessageRoleUser, SourceMessageID: "riido-comment-1", Body: "monolith already counts this"},
			{MessageID: "started", Role: AIAgentTaskThreadMessageRoleAgent, CommentKind: AgentTaskCommentAssignmentStarted, Body: "started"},
			{MessageID: "progress", Role: AIAgentTaskThreadMessageRoleProgress, Body: "reading files"},
			{MessageID: "empty", Role: AIAgentTaskThreadMessageRoleAgent, Body: "  "},
			{MessageID: "answer", Role: AIAgentTaskThreadMessageRoleAgent, CommentKind: AgentTaskCommentTaskCompleted, ResultMessage: "done"},
		}},
		{ThreadID: "standalone", Messages: []AIAgentTaskThreadHistoryMessage{
			{MessageID: "typed-reply", Role: AIAgentTaskThreadMessageRoleUser, SourceMessageID: "thread-message:receipt", Body: "follow-up"},
			{MessageID: "external", Role: AIAgentTaskThreadMessageRoleUser, Body: "from Slack"},
			{MessageID: "question", Role: AIAgentTaskThreadMessageRoleAgent, CommentKind: AgentTaskCommentNeedsInput, Body: "which branch?"},
		}},
		{ThreadID: "rerun", Messages: []AIAgentTaskThreadHistoryMessage{
			{MessageID: "answer", Role: AIAgentTaskThreadMessageRoleAgent, ResultMessage: "done"},
		}},
	}
	if got := countVisibleAgentCommentMessages(threads); got != 4 {
		t.Fatalf("count = %d, want answer, typed-reply, external, question", got)
	}
}
