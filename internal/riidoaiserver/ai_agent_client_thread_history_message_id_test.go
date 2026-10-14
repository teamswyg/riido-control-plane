package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTaskThreadHistoryMessageIDsMatchLegacyKeys(t *testing.T) {
	observedAt := time.Date(2026, 7, 2, 14, 1, 2, 3, time.UTC)
	assertHistoryID(t,
		taskThreadUserMessageID("thread-a", "asn-1", "", "body", observedAt),
		"msg_user_",
		strings.Join([]string{"thread-a", "asn-1", "body", observedAt.Format(time.RFC3339Nano)}, "\x00"),
	)
	assertHistoryID(t,
		taskThreadUserMessageID("thread-a", "asn-1", " source-1 ", "body", observedAt),
		"msg_user_",
		"source-1",
	)
	thread := AIAgentTaskThreadRecord{
		ThreadID: "thread-a", AssignmentID: "asn-1", RunID: "run-1",
		CommentKind: AgentTaskCommentTaskCompleted, AssignmentState: AgentAssignmentStateCompleted,
	}
	assertHistoryID(t, taskThreadProjectionMessageID(thread), "msg_agent_", strings.Join([]string{
		thread.ThreadID,
		thread.AssignmentID,
		thread.RunID,
		string(thread.CommentKind),
		string(thread.AssignmentState),
	}, "\x00"))
	assertHistoryID(t,
		taskThreadProgressMessageID("thread-a", 42),
		"msg_progress_",
		strings.Join([]string{"thread-a", "progress", strconv.Itoa(42)}, "\x00"),
	)
}

func TestTaskThreadProgressMessageIDFallsBackForLongThreadID(t *testing.T) {
	threadID := strings.Repeat("thread-", 120)
	assertHistoryID(t,
		taskThreadProgressMessageID(threadID, -7),
		"msg_progress_",
		strings.Join([]string{threadID, "progress", strconv.Itoa(-7)}, "\x00"),
	)
}

func assertHistoryID(t *testing.T, got, prefix, key string) {
	t.Helper()
	sum := sha256.Sum256([]byte(key))
	want := prefix + hex.EncodeToString(sum[:8])
	if got != want {
		t.Fatalf("message id = %q, want %q", got, want)
	}
}
