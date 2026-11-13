package riidoaiserver

import "strings"

func (s *DevelopmentAIAgentClientStore) repairCompletedTaskThreadProjectionLocked(thread AIAgentTaskThreadRecord, projection AssignmentProjection) bool {
	if projection.Assignment.State != AssignmentCompleted || !assignmentProjectionMatchesTaskThread(thread, projection) {
		return false
	}
	message, metadata := assignmentProjectionTerminalResult(projection)
	return s.repairCompletedTaskThreadResultLocked(thread, message, metadata)
}

func taskThreadNeedsCompletedResult(thread AIAgentTaskThreadRecord) bool {
	if thread.AssignmentState != AgentAssignmentStateCompleted || thread.WorkStatus != AgentWorkStatusCompleted ||
		thread.CommentKind != AgentTaskCommentTaskCompleted {
		return false
	}
	for _, value := range []string{thread.Message, thread.ResultMessage} {
		value = clientVisibleTaskThreadText(strings.TrimSpace(value))
		if value != "" && value != clientMessageTaskCompleted {
			return false
		}
	}
	return true
}

// Repair only the missing result of the same closed execution. Agent status may
// already belong to another task, and the original completion timestamp is kept.
func (s *DevelopmentAIAgentClientStore) repairCompletedTaskThreadResultLocked(thread AIAgentTaskThreadRecord, message string, metadata map[string]string) bool {
	if !taskThreadNeedsCompletedResult(thread) {
		return false
	}
	response := assignmentEventActionResponse(thread, AssignmentCompleted, message, metadata)
	if clientVisibleTaskThreadText(response.ResultMessage) == "" ||
		(clientVisibleTaskThreadText(response.ResultMessage) == clientMessageTaskCompleted && !assignmentMetadataNeedsInput(metadata)) {
		return false
	}
	threads := s.taskThreads[thread.TaskID]
	for i := range threads {
		if threads[i].ThreadID != thread.ThreadID || threads[i].AssignmentID != thread.AssignmentID || !taskThreadNeedsCompletedResult(threads[i]) {
			continue
		}
		threads[i].Message, threads[i].ResultMessage = response.Message, response.ResultMessage
		threads[i].WorkStatus, threads[i].AssignmentState = response.WorkStatus, response.AssignmentState
		threads[i].CommentKind, threads[i].ConfirmationReason = response.CommentKind, response.ConfirmationReason
		threads[i].FailureDiagnostics = copyFailureDiagnostics(response.FailureDiagnostics)
		threads[i].ExecutionContext = copyExecutionContext(response.ExecutionContext)
		s.taskThreads[thread.TaskID] = threads
		s.appendAgentTaskActionEvent(response)
		return true
	}
	return false
}
