package riidoaiserver

import (
	"strings"
	"time"
)

func applyTaskThreadStopProjection(thread *AIAgentTaskThreadRecord, response AIAgentTaskActionResponse, completed bool, now time.Time) {
	thread.WorkStatus = response.WorkStatus
	thread.AssignmentState = response.AssignmentState
	thread.CommentKind = response.CommentKind
	thread.Message = response.Message
	thread.ResultMessage = response.ResultMessage
	if completed {
		if strings.TrimSpace(thread.ResultMessage) == "" {
			thread.ResultMessage = response.Message
		}
		thread.CompletedAt = now
		return
	}
	thread.ResultMessage = ""
	thread.CompletedAt = time.Time{}
}
