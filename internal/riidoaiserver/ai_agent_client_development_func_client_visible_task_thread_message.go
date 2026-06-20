package riidoaiserver

import (
	"slices"
)

func clientVisibleTaskThreadMessage(thread AIAgentTaskThreadRecord) string {
	if message := clientVisibleTaskThreadText(thread.Message); message != "" {
		return message
	}
	for _, v := range slices.Backward(thread.Lines) {
		if message := clientVisibleTaskThreadText(v.Message); message != "" {
			return message
		}
	}
	switch thread.CommentKind {
	case AgentTaskCommentTaskCompleted:
		return "agent work completed"
	case AgentTaskCommentTaskFailed:
		return "agent work failed"
	case AgentTaskCommentStoppedByUserRequest, AgentTaskCommentStoppedByAgentDeleted:
		return "agent work was stopped"
	case AgentTaskCommentQueuedByBusyAgent:
		return "agent assignment is queued"
	case AgentTaskCommentAssignmentStarted, AgentTaskCommentRuntimeProgress:
		return "agent work is running"
	default:
		return ""
	}
}
