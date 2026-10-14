package riidoaiserver

import "strings"

func taskThreadConversationID(thread AIAgentTaskThreadRecord) string {
	if id := strings.TrimSpace(thread.ConversationID); id != "" {
		return id
	}
	return strings.TrimSpace(thread.ThreadID)
}

func taskThreadParentThreadID(thread AIAgentTaskThreadRecord) string {
	return strings.TrimSpace(thread.ParentThreadID)
}
