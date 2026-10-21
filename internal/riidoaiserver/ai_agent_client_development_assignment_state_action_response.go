package riidoaiserver

func applyAssignmentStateActionResponse(response *AIAgentTaskActionResponse, state AssignmentState) {
	switch state.Code() {
	case AssignmentStateCodeQueued, AssignmentStateCodeLeased:
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		ensureAssignmentResponseMessage(response, clientMessageTaskQueued)
	case AssignmentStateCodeReady:
		response.WorkStatus = AgentWorkStatusRunning
		response.AssignmentState = AgentAssignmentStateRunning
		response.CommentKind = AgentTaskCommentAssignmentStarted
		ensureAssignmentRunningResponseMessage(response)
	case AssignmentStateCodeRunning:
		response.WorkStatus = AgentWorkStatusRunning
		response.AssignmentState = AgentAssignmentStateRunning
		response.CommentKind = AgentTaskCommentRuntimeProgress
		ensureAssignmentRunningResponseMessage(response)
	case AssignmentStateCodeCancelling:
		response.WorkStatus = AgentWorkStatusIdle
		response.AssignmentState = AgentAssignmentStateStopped
		response.CommentKind = AgentTaskCommentStoppedByUserRequest
		ensureAssignmentResponseMessage(response, clientMessageTaskStopped)
	case AssignmentStateCodeCancelled:
		response.WorkStatus = AgentWorkStatusIdle
		response.AssignmentState = AgentAssignmentStateStopped
		response.CommentKind = AgentTaskCommentStoppedByUserRequest
		ensureAssignmentResponseMessage(response, clientMessageTaskStopped)
	case AssignmentStateCodeCompleted:
		response.WorkStatus = AgentWorkStatusCompleted
		response.AssignmentState = AgentAssignmentStateCompleted
		response.CommentKind = AgentTaskCommentTaskCompleted
		ensureAssignmentResponseMessage(response, clientMessageTaskCompleted)
	case AssignmentStateCodeFailed:
		response.WorkStatus = AgentWorkStatusFailed
		response.AssignmentState = AgentAssignmentStateFailed
		response.CommentKind = AgentTaskCommentTaskFailed
		ensureAssignmentResponseMessage(response, clientMessageTaskFailed)
	default:
		ensureAssignmentResponseMessage(response, clientMessageTaskRunning)
	}
}

func ensureAssignmentResponseMessage(response *AIAgentTaskActionResponse, fallback string) {
	if response.Message == "" {
		response.Message = fallback
	}
}

// Queue status is not provider output and must not become the next run's progress.
func ensureAssignmentRunningResponseMessage(response *AIAgentTaskActionResponse) {
	if assignmentMessageIsQueueStatus(response.Message) {
		response.Message = ""
	}
	ensureAssignmentResponseMessage(response, clientMessageTaskRunning)
}

func assignmentMessageIsQueueStatus(message string) bool {
	switch clientVisibleTaskThreadText(message) {
	case clientMessageAgentBusyQueued, clientMessageTaskQueued:
		return true
	default:
		return false
	}
}
