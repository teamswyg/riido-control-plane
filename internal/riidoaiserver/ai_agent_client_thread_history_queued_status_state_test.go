package riidoaiserver

import (
	"testing"
	"time"
)

func TestAIAgentTaskThreadHistoryHidesQueuedMessageWhenConversationIsRunning(t *testing.T) {
	queuedAt := time.Unix(10, 0).UTC()
	runningAt := time.Unix(11, 0).UTC()
	threads := []AIAgentTaskThreadHistoryRecord{
		historyRecordWithMessages("thread-queued", "conversation-a", []AIAgentTaskThreadHistoryMessage{
			historyUserMessage("이어서 해줘", queuedAt),
			historyAgentMessage(AgentTaskCommentQueuedByBusyAgent, queuedAt),
		}),
		{
			ThreadID:        "thread-running",
			ConversationID:  "conversation-a",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			StartedAt:       runningAt,
		},
	}
	suppressSupersededQueuedHistoryMessages(threads)
	if historyHasAgentCommentKind(threads[0].Messages, AgentTaskCommentQueuedByBusyAgent) {
		t.Fatalf("running conversation must hide stale queued message: %+v", threads[0].Messages)
	}
	if !historyMessagesContainUserBody(threads[0].Messages, "이어서 해줘") {
		t.Fatalf("user follow-up must remain visible: %+v", threads[0].Messages)
	}
}
