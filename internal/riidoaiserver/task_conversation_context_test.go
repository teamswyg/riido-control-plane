package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestFollowupUsesOnlySelectedConversationAndLatestDirective(t *testing.T) {
	now := time.Unix(1, 0)
	msg := func(id, role, body string, offset int) AIAgentTaskThreadHistoryMessage {
		return AIAgentTaskThreadHistoryMessage{MessageID: id, Role: AIAgentTaskThreadMessageRole(role), Body: body, ObservedAt: now.Add(time.Duration(offset) * time.Second)}
	}
	current := AIAgentTaskThreadRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent", ConversationID: "conversation"}
	owned := AIAgentTaskThreadHistoryRecord{ThreadID: "old", TaskID: "task", AgentID: "agent", ConversationID: "conversation", Messages: []AIAgentTaskThreadHistoryMessage{msg("b", "agent", "Implemented auth.go and passed test.", 2), msg("a", "user", "<p>Fix auth.go</p>", 1), msg("c", "progress", "internal-progress-only", 3)}}
	other := owned
	other.ConversationID = "other"
	other.Messages = []AIAgentTaskThreadHistoryMessage{msg("d", "user", "other-conversation-private", 4)}
	foreign := owned
	foreign.TaskID = "other-task"
	foreign.Messages = []AIAgentTaskThreadHistoryMessage{msg("e", "user", "other-task-private", 4)}
	wrongAgent := owned
	wrongAgent.AgentID = "other-agent"
	wrongAgent.Messages = []AIAgentTaskThreadHistoryMessage{msg("f", "user", "other-agent-private", 4)}
	got := appendAIAgentTaskThreadMessagePrompt("current task snapshot", current, CreateAIAgentTaskThreadMessageRequest{Body: "<p>Now add a regression test.</p>"}, owned, other, foreign, wrongAgent, owned)
	for _, want := range []string{"current task snapshot", "Fix auth.go", "Implemented auth.go", "Now add a regression test."} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, bad := range []string{"other-conversation-private", "other-task-private", "other-agent-private", "internal-progress-only", "<p>"} {
		if strings.Contains(got, bad) {
			t.Errorf("leaked %q", bad)
		}
	}
	if strings.Count(got, "Fix auth.go") != 1 || strings.Index(got, "Fix auth.go") > strings.Index(got, "Implemented auth.go") || strings.Index(got, "Implemented auth.go") > strings.Index(got, "Now add") {
		t.Fatal("history order/dedup/latest directive invalid")
	}
}

func TestFollowupBudgetKeepsRecentHistoryAndRejectsCrossTaskSelection(t *testing.T) {
	thread := AIAgentTaskThreadRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent"}
	record := AIAgentTaskThreadHistoryRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent"}
	for i := range 20 {
		record.Messages = append(record.Messages, AIAgentTaskThreadHistoryMessage{Role: AIAgentTaskThreadMessageRoleAgent, Body: strings.Repeat("content ", 1000), ObservedAt: time.Unix(int64(i), 0)})
	}
	var out strings.Builder
	appendTaskConversationContext(&out, thread, CreateAIAgentTaskThreadMessageRequest{}, []AIAgentTaskThreadHistoryRecord{record})
	if len([]rune(out.String())) > conversationContextLimit+300 || !strings.Contains(out.String(), "Earlier conversation messages omitted") {
		t.Fatal("history budget missing")
	}
	_, err := selectAIAgentHistoryThreadForFollowup(AIAgentTaskThreadHistoryCollectionResponse{TaskID: "other-task", Threads: []AIAgentTaskThreadHistoryRecord{record}}, "task", "thread")
	if err == nil {
		t.Fatal("foreign task history admitted")
	}
}
