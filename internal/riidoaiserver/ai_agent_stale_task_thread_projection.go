package riidoaiserver

import "time"

func applyStaleActiveTaskThreadProjection(thread *AIAgentTaskThreadRecord, now time.Time) {
	if thread.AssignmentState == AgentAssignmentStateStopping {
		applyStaleStoppingTaskThreadProjection(thread, now)
		return
	}
	message := clientMessageTaskTimeout
	thread.WorkStatus = AgentWorkStatusFailed
	thread.AssignmentState = AgentAssignmentStateFailed
	thread.CommentKind = AgentTaskCommentTaskFailed
	thread.Message = message
	thread.ResultMessage = message
	thread.QueueDiagnostics = nil
	thread.FailureDiagnostics = &AIAgentTaskThreadFailureDiagnostics{
		ResultStatus:    "failed",
		FailureCategory: "stale_active_thread_timeout",
		Message:         message,
	}
	thread.CompletedAt = now
}

func applyStaleStoppingTaskThreadProjection(thread *AIAgentTaskThreadRecord, now time.Time) {
	message := clientMessageTaskTimeout
	thread.WorkStatus = AgentWorkStatusIdle
	thread.AssignmentState = AgentAssignmentStateStopped
	thread.CommentKind = AgentTaskCommentStoppedByUserRequest
	thread.Message = message
	thread.ResultMessage = message
	thread.QueueDiagnostics = nil
	thread.FailureDiagnostics = nil
	thread.CompletedAt = now
}
