package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) updateTaskThreadMessageFromActionLocked(
	thread *AIAgentTaskThreadRecord,
	response AIAgentTaskActionResponse,
	sourceMessageID string,
	conversationID string,
	parentThreadID string,
	now time.Time,
) {
	thread.RunID = response.RunID
	if thread.ConversationID == "" {
		thread.ConversationID = defaultTaskThreadConversationID(conversationID, response.ThreadID)
	}
	if parent := strings.TrimSpace(parentThreadID); parent != "" {
		thread.ParentThreadID = parent
	}
	if assignmentID := strings.TrimSpace(response.AssignmentID); assignmentID != "" {
		thread.AssignmentID = assignmentID
	}
	s.updateTaskThreadMessageAgentSnapshotLocked(thread, response, now)
	thread.WorkStatus = response.WorkStatus
	thread.AssignmentState = response.AssignmentState
	thread.FailureDiagnostics = copyFailureDiagnostics(response.FailureDiagnostics)
	thread.CommentKind = response.CommentKind
	thread.Message = response.Message
	thread.ResultMessage = response.ResultMessage
	if sourceMessageID != "" {
		thread.SourceMessageID = sourceMessageID
	}
	if thread.StartedAt.IsZero() {
		thread.StartedAt = now
	}
	updateTaskThreadMessageCompletedAt(thread, now)
}

func updateTaskThreadMessageCompletedAt(thread *AIAgentTaskThreadRecord, now time.Time) {
	if taskThreadHasActiveStream(*thread) {
		thread.CompletedAt = time.Time{}
		return
	}
	thread.CompletedAt = now
}
