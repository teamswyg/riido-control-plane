package riidoaiserver

import (
	"fmt"
	"testing"
)

func chunkedThreadSnapshotForTest(t *testing.T) AIAgentClientSnapshot {
	t.Helper()
	snapshot := conversationSnapshotForTest(200)
	threads := make([]AIAgentTaskThreadRecord, 200)
	for i, message := range snapshot.TaskThreadMessages["thread"] {
		threads[i] = AIAgentTaskThreadRecord{
			TaskID: "task", ThreadID: fmt.Sprintf("thread-%d", i), AgentID: "agent-a",
			ConversationID: "comment:root", Message: message.Body, ResultMessage: message.Body,
			AgentSnapshot: &AIAgentTaskThreadAgentSnapshot{AgentID: "agent-a", Name: "Agent A", CapturedAt: fixedSnapshotTestNow(), ProfileColor: fallbackAgentProfileColor("agent-a")},
			Lines:         []AgentThreadProgressLine{},
			WorkStatus:    AgentWorkStatusCompleted, AssignmentState: AgentAssignmentStateCompleted,
		}
	}
	snapshot.TaskThreads["task"] = threads
	part, err := encodeDynamoDBAIAgentClientSnapshotPart(dynamoDBAIAgentClientSnapshotPartTaskThreads, snapshot.TaskThreads)
	if err != nil || len(part.gzip) <= conversationChunkBytes {
		t.Fatalf("fixture must split TASK_THREADS: bytes=%d err=%v", len(part.gzip), err)
	}
	return snapshot
}
