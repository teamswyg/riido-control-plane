package riidoaiserver

import (
	"reflect"
	"testing"
	"time"
)

func TestAIAgentTaskThreadHistorySortsMessagesDeterministically(t *testing.T) {
	at := time.Date(2026, 6, 25, 4, 0, 0, 0, time.UTC)
	messages := []AIAgentTaskThreadHistoryMessage{
		{MessageID: "agent", Role: AIAgentTaskThreadMessageRoleAgent, ObservedAt: at},
		{MessageID: "progress-2", Role: AIAgentTaskThreadMessageRoleProgress, Seq: 2, ObservedAt: at},
		{MessageID: "user", Role: AIAgentTaskThreadMessageRoleUser, ObservedAt: at},
		{MessageID: "progress-1", Role: AIAgentTaskThreadMessageRoleProgress, Seq: 1, ObservedAt: at},
	}
	sortTaskThreadHistoryMessages(messages)
	got := historyMessageIDs(messages)
	want := []string{"user", "progress-1", "progress-2", "agent"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("message order = %v, want %v", got, want)
	}
}

func historyMessageIDs(messages []AIAgentTaskThreadHistoryMessage) []string {
	out := make([]string, 0, len(messages))
	for _, message := range messages {
		out = append(out, message.MessageID)
	}
	return out
}
