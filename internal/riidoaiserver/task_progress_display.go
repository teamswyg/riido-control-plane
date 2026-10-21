package riidoaiserver

import (
	"strings"
	"unicode/utf8"
)

// Project known provider-authored descriptions, never infer task activity from
// raw transcript text or expose arbitrary metadata in client responses.
func taskProgressDisplayMessage(line AgentThreadProgressLine) string {
	if line.displayMessage != "" && line.MessageCode == 0 && line.MessageKey == "" {
		return line.displayMessage
	}
	var key, arg string
	switch line.MessageCode {
	case 1101:
		key, arg = "tool.collecting", "description"
	case 1103:
		key, arg = "tool.running", "description"
	case 1104:
		key, arg = "tool.completed", "summary"
	default:
		return ""
	}
	if line.MessageKey != "" && line.MessageKey != key {
		return ""
	}
	text := clientVisibleTaskThreadText(line.MessageArgs[arg])
	if text == "" || strings.ContainsAny(text, "\r\n") || utf8.RuneCountInString(text) > 180 {
		return ""
	}
	return text
}
