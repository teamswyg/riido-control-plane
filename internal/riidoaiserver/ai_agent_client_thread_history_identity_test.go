package riidoaiserver

import "testing"

func TestDefaultTaskThreadConversationID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		conversationID string
		threadID       string
		want           string
	}{
		{name: "explicit conversation", conversationID: " conv-1 ", threadID: "thread-1", want: "conv-1"},
		{name: "thread fallback", conversationID: "   ", threadID: " thread-2 ", want: "thread-2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := defaultTaskThreadConversationID(tt.conversationID, tt.threadID); got != tt.want {
				t.Fatalf("conversation id = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskThreadMessageRunID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		taskID       string
		assignmentID string
		sequence     int
		want         string
	}{
		{name: "assignment scoped", taskID: "task-1", assignmentID: "asn-1", sequence: 7, want: "run-dev-message-task-1-asn-1"},
		{name: "sequence fallback", taskID: "task-1", sequence: 7, want: "run-dev-message-task-1-7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := taskThreadMessageRunID(tt.taskID, tt.assignmentID, tt.sequence); got != tt.want {
				t.Fatalf("run id = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRetainLatestThreadHistoryMessagesCopiesAndTrims(t *testing.T) {
	t.Parallel()
	messages := make([]AIAgentTaskThreadHistoryMessage, aiAgentClientThreadHistoryMessageLimit+2)
	for i := range messages {
		messages[i] = AIAgentTaskThreadHistoryMessage{MessageID: "msg", Seq: i + 1, Body: "body"}
	}
	retained := retainLatestThreadHistoryMessages(messages)
	if len(retained) != aiAgentClientThreadHistoryMessageLimit {
		t.Fatalf("len(retained) = %d, want %d", len(retained), aiAgentClientThreadHistoryMessageLimit)
	}
	if retained[0].Seq != 3 || retained[len(retained)-1].Seq != aiAgentClientThreadHistoryMessageLimit+2 {
		t.Fatalf("retained seq range = %d..%d, want 3..%d", retained[0].Seq, retained[len(retained)-1].Seq, aiAgentClientThreadHistoryMessageLimit+2)
	}
	messages[2].Body = "mutated"
	if retained[0].Body != "body" {
		t.Fatalf("retained message aliases source body = %q", retained[0].Body)
	}
}
