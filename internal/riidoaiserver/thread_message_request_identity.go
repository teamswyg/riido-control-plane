package riidoaiserver

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func threadMessageRequestID(threadID, requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ""
	}
	return fmt.Sprintf("thread-message-%x", sha256.Sum256([]byte(strings.TrimSpace(threadID)+"\x00"+requestID)))
}
