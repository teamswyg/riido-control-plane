package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestResumedFollowupSendsOnlyNewConversationContext(t *testing.T) {
	now := time.Now().UTC()
	thread := AIAgentTaskThreadRecord{ThreadID: "mine", TaskID: "task", AgentID: "agent", ConversationID: "comment:root", StartedAt: now}
	history := []AIAgentTaskThreadHistoryRecord{
		{ThreadID: "mine", TaskID: "task", AgentID: "agent", ConversationID: "comment:root", Messages: []AIAgentTaskThreadHistoryMessage{{Role: AIAgentTaskThreadMessageRoleUser, Body: "known native request", SourceMessageID: "known", ObservedAt: now}}},
		{TaskID: "task", AgentID: "agent", ConversationID: "comment:root", Messages: []AIAgentTaskThreadHistoryMessage{
			{Role: AIAgentTaskThreadMessageRoleUser, Body: "old root context", ObservedAt: now.Add(-time.Minute)},
			{Role: AIAgentTaskThreadMessageRoleUser, Body: "known comment copy", SourceMessageID: "known", ObservedAt: now},
			{Role: AIAgentTaskThreadMessageRoleUser, Body: "new ordinary reply", ObservedAt: now.Add(time.Minute)},
			{Role: AIAgentTaskThreadMessageRoleUser, Body: "current instruction duplicate", SourceMessageID: "current", ObservedAt: now.Add(time.Minute)},
		}},
		{ThreadID: "other", TaskID: "task", AgentID: "other-agent", ConversationID: "comment:root", Messages: []AIAgentTaskThreadHistoryMessage{{Role: AIAgentTaskThreadMessageRoleAgent, Body: "new specialist finding", ObservedAt: now.Add(time.Minute)}}},
		{ThreadID: "foreign", TaskID: "other-task", AgentID: "other-agent", ConversationID: "comment:root", Messages: []AIAgentTaskThreadHistoryMessage{{Role: AIAgentTaskThreadMessageRoleAgent, Body: "foreign secret", ObservedAt: now.Add(time.Minute)}}},
	}
	got := resumedFollowupPrompt("<p>Continue with the final revision.</p>", thread, CreateAIAgentTaskThreadMessageRequest{SourceMessageID: "current"}, history)
	for _, want := range []string{"Continue with the final revision.", "new ordinary reply", "new specialist finding", "riido context get"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q: %s", want, got)
		}
	}
	for _, unwanted := range []string{"known native request", "old root context", "known comment copy", "current instruction duplicate", "foreign secret", "Riido AI Agent Assignment"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("repeated or leaked %q", unwanted)
		}
	}
}
