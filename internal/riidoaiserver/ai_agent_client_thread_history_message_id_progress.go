package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

const (
	taskThreadProgressMessagePrefix = "msg_progress_"
	taskThreadProgressKeyPart       = "progress"
)

func taskThreadProgressMessageIDFast(threadID string, seq int) string {
	var key [512]byte
	n, ok := appendTaskThreadProgressKeyStack(key[:0], threadID, seq)
	if !ok {
		return taskThreadHistoryHashID(
			taskThreadProgressMessagePrefix,
			taskThreadProgressMessageKey(threadID, seq),
		)
	}
	sum := sha256.Sum256(key[:n])
	var out [len(taskThreadProgressMessagePrefix) + 16]byte
	copy(out[:], taskThreadProgressMessagePrefix)
	hex.Encode(out[len(taskThreadProgressMessagePrefix):], sum[:8])
	return string(out[:])
}

func appendTaskThreadProgressKeyStack(buf []byte, threadID string, seq int) (int, bool) {
	needed := len(threadID) + len(taskThreadProgressKeyPart) + 2 + 20
	if cap(buf) < needed {
		return 0, false
	}
	buf = append(buf, threadID...)
	buf = append(buf, 0)
	buf = append(buf, taskThreadProgressKeyPart...)
	buf = append(buf, 0)
	buf = strconv.AppendInt(buf, int64(seq), 10)
	return len(buf), true
}
