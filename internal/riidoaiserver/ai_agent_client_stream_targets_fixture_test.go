package riidoaiserver

import (
	"strconv"
	"time"
)

func streamTargetFixtureStore(threadCount, lineCount int) *DevelopmentAIAgentClientStore {
	store := NewDevelopmentAIAgentClientStore()
	now := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	threads := make([]AIAgentTaskThreadRecord, threadCount)
	for i := range threads {
		threads[i] = AIAgentTaskThreadRecord{
			ThreadID:        "thread-load-" + strconv.Itoa(i),
			TaskID:          "task-load-read",
			AssignmentID:    "asn-load-" + strconv.Itoa(i),
			AgentID:         "agent-owned-codex",
			RunID:           "run-load-" + strconv.Itoa(i),
			WorkStatus:      AgentWorkStatusRunning,
			AssignmentState: AgentAssignmentStateRunning,
			CommentKind:     AgentTaskCommentRuntimeProgress,
			StartedAt:       now,
			Lines:           streamTargetFixtureLines(lineCount),
		}
	}
	store.taskThreads["task-load-read"] = threads
	return store
}

func streamTargetFixtureLines(count int) []AgentThreadProgressLine {
	lines := make([]AgentThreadProgressLine, count)
	for i := range lines {
		lines[i] = AgentThreadProgressLine{Seq: i + 1, Message: "line " + strconv.Itoa(i)}
	}
	return lines
}
