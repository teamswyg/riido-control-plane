package riidoaiserver

import (
	"testing"
	"time"
)

func TestAIAgentTaskThreadHistoryHidesSupersededBusyQueuedMessage(t *testing.T) {
	queuedAt := time.Unix(10, 0).UTC()
	runningAt := time.Unix(11, 0).UTC()
	threads := []AIAgentTaskThreadHistoryRecord{
		historyRecordWithMessages("thread-queued", "conversation-a", []AIAgentTaskThreadHistoryMessage{
			historyUserMessage("먼저 기다려줘", queuedAt),
			historyAgentMessage(AgentTaskCommentQueuedByBusyAgent, queuedAt),
		}),
		historyRecordWithMessages("thread-running", "conversation-a", []AIAgentTaskThreadHistoryMessage{
			historyAgentMessage(AgentTaskCommentRuntimeProgress, runningAt),
		}),
	}
	suppressSupersededQueuedHistoryMessages(threads)
	if historyHasAgentCommentKind(threads[0].Messages, AgentTaskCommentQueuedByBusyAgent) {
		t.Fatalf("superseded queued message leaked into v3 history: %+v", threads[0].Messages)
	}
	if !historyMessagesContainUserBody(threads[0].Messages, "먼저 기다려줘") {
		t.Fatalf("queued user follow-up must remain in history: %+v", threads[0].Messages)
	}
}

func TestAIAgentTaskThreadHistoryKeepsCurrentBusyQueuedMessage(t *testing.T) {
	queuedAt := time.Unix(10, 0).UTC()
	threads := []AIAgentTaskThreadHistoryRecord{
		historyRecordWithMessages("thread-queued", "conversation-a", []AIAgentTaskThreadHistoryMessage{
			historyAgentMessage(AgentTaskCommentQueuedByBusyAgent, queuedAt),
		}),
	}
	suppressSupersededQueuedHistoryMessages(threads)
	if !historyHasAgentCommentKind(threads[0].Messages, AgentTaskCommentQueuedByBusyAgent) {
		t.Fatalf("current queued message must remain visible: %+v", threads[0].Messages)
	}
}

func historyRecordWithMessages(
	threadID string,
	conversationID string,
	messages []AIAgentTaskThreadHistoryMessage,
) AIAgentTaskThreadHistoryRecord {
	return AIAgentTaskThreadHistoryRecord{
		ThreadID: threadID, ConversationID: conversationID, Messages: messages,
	}
}

func historyAgentMessage(kind AgentTaskCommentKind, observedAt time.Time) AIAgentTaskThreadHistoryMessage {
	return AIAgentTaskThreadHistoryMessage{
		MessageID: string(kind), Role: AIAgentTaskThreadMessageRoleAgent,
		CommentKind: kind, Body: string(kind), ObservedAt: observedAt,
	}
}

func historyUserMessage(body string, observedAt time.Time) AIAgentTaskThreadHistoryMessage {
	return AIAgentTaskThreadHistoryMessage{
		MessageID: "user-" + body, Role: AIAgentTaskThreadMessageRoleUser,
		Body: body, ObservedAt: observedAt,
	}
}

func historyHasAgentCommentKind(messages []AIAgentTaskThreadHistoryMessage, kind AgentTaskCommentKind) bool {
	for _, message := range messages {
		if message.Role == AIAgentTaskThreadMessageRoleAgent && message.CommentKind == kind {
			return true
		}
	}
	return false
}
