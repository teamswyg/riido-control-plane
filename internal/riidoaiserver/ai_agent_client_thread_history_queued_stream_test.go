package riidoaiserver

import (
	"testing"
	"time"
)

func TestAIAgentTaskThreadHistoryClearsQueuedStreamWhenConversationRuns(t *testing.T) {
	queuedAt := time.Unix(11, 0).UTC()
	runningAt := time.Unix(10, 0).UTC()
	threads := []AIAgentTaskThreadHistoryRecord{
		{
			ThreadID:        "thread-running",
			ConversationID:  "conversation-a",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			StartedAt:       runningAt,
			ActiveStream:    historyStream("thread-running", "run-running"),
		},
		{
			ThreadID:        "thread-queued",
			ConversationID:  "conversation-a",
			WorkStatus:      AgentWorkStatusQueued,
			AssignmentState: AgentAssignmentStateQueued,
			StartedAt:       queuedAt,
			ActiveStream:    historyStream("thread-queued", "run-queued"),
			Messages: []AIAgentTaskThreadHistoryMessage{
				historyAgentMessage(AgentTaskCommentQueuedByBusyAgent, queuedAt),
			},
		},
	}
	suppressSupersededQueuedHistoryMessages(threads)
	if threads[1].ActiveStream != nil {
		t.Fatalf("superseded queued thread kept active stream: %+v", threads[1].ActiveStream)
	}
	if historyHasAgentCommentKind(threads[1].Messages, AgentTaskCommentQueuedByBusyAgent) {
		t.Fatalf("superseded queued message leaked: %+v", threads[1].Messages)
	}
}

func TestAIAgentTaskThreadHistoryActiveStreamPrefersRunning(t *testing.T) {
	threads := []AIAgentTaskThreadHistoryRecord{
		{
			ThreadID:        "thread-running",
			ConversationID:  "conversation-a",
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			ActiveStream:    historyStream("thread-running", "run-running"),
		},
		{
			ThreadID:        "thread-queued",
			ConversationID:  "conversation-b",
			WorkStatus:      AgentWorkStatusQueued,
			AssignmentState: AgentAssignmentStateQueued,
			ActiveStream:    historyStream("thread-queued", "run-queued"),
		},
	}
	stream := taskThreadHistoryActiveStream(threads)
	if stream == nil || stream.ThreadID != "thread-running" {
		t.Fatalf("active stream should prefer running over queued: %+v", stream)
	}
}

func historyStream(threadID, runID string) *AIAgentTaskThreadStreamLink {
	return &AIAgentTaskThreadStreamLink{ThreadID: threadID, RunID: runID}
}
