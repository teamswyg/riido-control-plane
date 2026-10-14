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

func TestAIAgentTaskThreadHistoryHidesCurrentBusyQueuedMessage(t *testing.T) {
	queuedAt := time.Unix(10, 0).UTC()
	threads := []AIAgentTaskThreadHistoryRecord{
		{
			ThreadID:        "thread-queued",
			ConversationID:  "conversation-a",
			WorkStatus:      AgentWorkStatusQueued,
			AssignmentState: AgentAssignmentStateQueued,
			Messages: []AIAgentTaskThreadHistoryMessage{
				historyAgentMessage(AgentTaskCommentQueuedByBusyAgent, queuedAt),
			},
		},
	}
	suppressSupersededQueuedHistoryMessages(threads)
	if historyHasAgentCommentKind(threads[0].Messages, AgentTaskCommentQueuedByBusyAgent) {
		t.Fatalf("queued status must not render as timeline message: %+v", threads[0].Messages)
	}
	if threads[0].AssignmentState != AgentAssignmentStateQueued ||
		threads[0].WorkStatus != AgentWorkStatusQueued {
		t.Fatalf("queued status must stay available on the thread: %+v", threads[0])
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
