package riidoaiserver

import (
	"testing"
	"time"
)

func TestAIAgentTaskThreadHistoryHidesQueuedMessageAfterRunningTimestamp(t *testing.T) {
	runningAt := time.Unix(10, 0).UTC()
	queuedAt := time.Unix(11, 0).UTC()
	threads := []AIAgentTaskThreadHistoryRecord{
		{
			ThreadID:        "thread-running",
			ConversationID:  "conversation-a",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			StartedAt:       runningAt,
		},
		historyRecordWithMessages("thread-queued", "conversation-a", []AIAgentTaskThreadHistoryMessage{
			historyUserMessage("다시 진행해줘", queuedAt),
			historyAgentMessage(AgentTaskCommentQueuedByBusyAgent, queuedAt),
		}),
	}
	suppressSupersededQueuedHistoryMessages(threads)
	if historyHasAgentCommentKind(threads[1].Messages, AgentTaskCommentQueuedByBusyAgent) {
		t.Fatalf("running conversation must hide queued message even when queued timestamp is later: %+v", threads[1].Messages)
	}
	if !historyMessagesContainUserBody(threads[1].Messages, "다시 진행해줘") {
		t.Fatalf("user follow-up must remain visible: %+v", threads[1].Messages)
	}
}
