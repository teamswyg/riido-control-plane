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
	if key != "" {
		return taskThreadHistoryHashID("msg_user_", taskThreadSourceMessageKey(key))
	}
	return taskThreadHistoryHashID("msg_user_", taskThreadUserMessageKey(threadID, assignmentID, body, observedAt))
}

func taskThreadProjectionMessageID(thread AIAgentTaskThreadRecord) string {
	return taskThreadHistoryHashID("msg_agent_", taskThreadProjectionMessageKey(thread))
}

func taskThreadProgressMessageID(threadID string, seq int) string {
	return taskThreadProgressMessageIDFast(threadID, seq)
}

func taskThreadHistoryHashID(prefix string, key []byte) string {
	sum := sha256.Sum256(key)
	out := make([]byte, len(prefix)+16)
	copy(out, prefix)
	hex.Encode(out[len(prefix):], sum[:8])
	return string(out)
}

func taskThreadUserMessageKey(threadID, assignmentID, body string, observedAt time.Time) []byte {
	observed := observedAt.Format(time.RFC3339Nano)
	key := make([]byte, 0, len(threadID)+len(assignmentID)+len(body)+len(observed)+3)
	key = appendHistoryKeyPart(key, threadID)
	key = appendHistoryKeyPart(key, assignmentID)
	key = appendHistoryKeyPart(key, body)
	return appendHistoryKeyPart(key, observed)
}

func taskThreadSourceMessageKey(sourceID string) []byte {
	key := make([]byte, 0, len(sourceID))
	return append(key, sourceID...)
}

func taskThreadProjectionMessageKey(thread AIAgentTaskThreadRecord) []byte {
	key := make([]byte, 0, len(thread.ThreadID)+len(thread.AssignmentID)+len(thread.RunID)+64)
	key = appendHistoryKeyPart(key, thread.ThreadID)
	key = appendHistoryKeyPart(key, thread.AssignmentID)
	key = appendHistoryKeyPart(key, thread.RunID)
	key = appendHistoryKeyPart(key, string(thread.CommentKind))
	return appendHistoryKeyPart(key, string(thread.AssignmentState))
}

func taskThreadProgressMessageKey(threadID string, seq int) []byte {
	key := make([]byte, 0, len(threadID)+16)
	key = appendHistoryKeyPart(key, threadID)
	key = appendHistoryKeyPart(key, "progress")
	if len(key) > 0 {
		key = append(key, 0)
	}
	return strconv.AppendInt(key, int64(seq), 10)
}

func appendHistoryKeyPart(key []byte, part string) []byte {
	if len(key) > 0 {
		key = append(key, 0)
	}
	return append(key, part...)
}
