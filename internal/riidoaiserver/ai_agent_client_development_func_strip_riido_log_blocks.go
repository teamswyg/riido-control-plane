package riidoaiserver

import (
	"strings"
)

const (
	riidoLogStart = "<riido_log>"
	riidoLogEnd   = "<end>"
)

func stripRiidoLogBlocks(message string) string {
	message = strings.TrimSpace(message)
	for {
		startIndex := strings.Index(message, riidoLogStart)
		if startIndex < 0 {
			return strings.TrimSpace(stripDanglingRiidoLogPrefix(message))
		}
		endOffset := strings.Index(message[startIndex+len(riidoLogStart):], riidoLogEnd)
		if endOffset < 0 {
			return strings.TrimSpace(message[:startIndex])
		}
		endIndex := startIndex + len(riidoLogStart) + endOffset + len(riidoLogEnd)
		message = message[:startIndex] + message[endIndex:]
	}
}

func stripDanglingRiidoLogPrefix(message string) string {
	for offset := 0; offset < len(message); offset++ {
		index := strings.Index(message[offset:], "<")
		if index < 0 {
			return message
		}
		index += offset
		if strings.HasPrefix(riidoLogStart, message[index:]) {
			return message[:index]
		}
		offset = index
	}
	return message
}
