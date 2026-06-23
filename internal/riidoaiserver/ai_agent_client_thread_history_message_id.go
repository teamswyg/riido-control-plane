package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func taskThreadUserMessageID(threadID, assignmentID, sourceID, body string, observedAt time.Time) string {
	key := strings.TrimSpace(sourceID)
	if key == "" {
		key = strings.Join([]string{threadID, assignmentID, body, observedAt.Format(time.RFC3339Nano)}, "\x00")
	}
	sum := sha256.Sum256([]byte(key))
	return "msg_user_" + hex.EncodeToString(sum[:8])
}

func taskThreadProjectionMessageID(thread AIAgentTaskThreadRecord) string {
	key := strings.Join([]string{
		thread.ThreadID,
		thread.AssignmentID,
		thread.RunID,
		string(thread.CommentKind),
		string(thread.AssignmentState),
	}, "\x00")
	sum := sha256.Sum256([]byte(key))
	return "msg_agent_" + hex.EncodeToString(sum[:8])
}

func taskThreadProgressMessageID(threadID string, seq int) string {
	key := strings.Join([]string{threadID, "progress", strconv.Itoa(seq)}, "\x00")
	sum := sha256.Sum256([]byte(key))
	return "msg_progress_" + hex.EncodeToString(sum[:8])
}
