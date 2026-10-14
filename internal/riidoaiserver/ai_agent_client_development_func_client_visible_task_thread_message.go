package riidoaiserver

import (
	"slices"
)

func clientVisibleTaskThreadMessage(thread AIAgentTaskThreadRecord) string {
	if thread.CommentKind == AgentTaskCommentStoppedByAgentDeleted {
		return clientMessageAgentDeleted
	}
	if message := clientVisibleTaskThreadText(thread.Message); message != "" {
		return message
	}
	for _, v := range slices.Backward(thread.Lines) {
		if message := clientVisibleTaskThreadText(v.Message); message != "" {
			return message
		}
	}
	return clientVisibleTaskThreadFallback(thread.CommentKind)
}
