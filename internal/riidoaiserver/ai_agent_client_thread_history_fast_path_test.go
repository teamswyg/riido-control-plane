package riidoaiserver

import "testing"

func TestAIAgentTaskThreadHistoryProgressOnlyUsesCachedMessages(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	thread := AIAgentTaskThreadRecord{
		ThreadID:     "thread-1",
		AssignmentID: "assignment-1",
		RunID:        "run-1",
		Lines:        []AgentThreadProgressLine{{Seq: 1, Message: "progress"}},
	}
	store.mu.Lock()
	first := store.taskThreadHistoryMessagesLocked(thread)
	second := store.taskThreadHistoryMessagesLocked(thread)
	store.mu.Unlock()
	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("progress history messages = %d/%d, want 2/2", len(first), len(second))
	}
	if &first[0] != &second[0] {
		t.Fatal("progress history should reuse cached messages")
	}
}

func TestWithoutSupersededQueuedMessagesKeepsNonQueuedSlice(t *testing.T) {
	messages := []AIAgentTaskThreadHistoryMessage{{
		MessageID: "progress-1",
		Role:      AIAgentTaskThreadMessageRoleProgress,
	}}
	got := withoutSupersededQueuedMessages("conversation-1", messages, nil, nil)
	if len(got) != 1 || &got[0] != &messages[0] {
		t.Fatalf("non-queued messages should remain unchanged: %+v", got)
	}
}
