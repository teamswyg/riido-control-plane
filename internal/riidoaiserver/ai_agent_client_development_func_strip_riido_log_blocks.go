package riidoaiserver

import (
	"strings"
)

func stripRiidoLogBlocks(message string) string {
	message = strings.TrimSpace(message)
	const start = "<riido_log>"
	const end = "<end>"
	for {
		startIndex := strings.Index(message, start)
		if startIndex < 0 {
			return strings.TrimSpace(message)
		}
		endOffset := strings.Index(message[startIndex+len(start):], end)
		if endOffset < 0 {
			return strings.TrimSpace(message)
		}
		endIndex := startIndex + len(start) + endOffset + len(end)
		message = message[:startIndex] + message[endIndex:]
	}
}
