package riidoaiserver

func applyAssignmentStateActionResponse(response *AIAgentTaskActionResponse, state AssignmentState) {
	switch state.Code() {
	case AssignmentStateCodeQueued, AssignmentStateCodeLeased:
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		ensureAssignmentResponseMessage(response, "agent assignment is queued")
	case AssignmentStateCodeReady:
		response.WorkStatus = AgentWorkStatusRunning
		response.AssignmentState = AgentAssignmentStateRunning
		response.CommentKind = AgentTaskCommentAssignmentStarted
		ensureAssignmentResponseMessage(response, "agent assignment was accepted by runtime")
	case AssignmentStateCodeRunning:
		response.WorkStatus = AgentWorkStatusRunning
		response.AssignmentState = AgentAssignmentStateRunning
		response.CommentKind = AgentTaskCommentRuntimeProgress
		ensureAssignmentResponseMessage(response, "agent work is running")
	case AssignmentStateCodeCancelling:
		response.WorkStatus = AgentWorkStatusRunning
		response.AssignmentState = AgentAssignmentStateStopping
		response.CommentKind = AgentTaskCommentStoppedByUserRequest
		ensureAssignmentResponseMessage(response, "agent work is stopping")
	case AssignmentStateCodeCancelled:
		response.WorkStatus = AgentWorkStatusIdle
		response.AssignmentState = AgentAssignmentStateStopped
		response.CommentKind = AgentTaskCommentStoppedByUserRequest
		ensureAssignmentResponseMessage(response, "agent work was stopped")
	case AssignmentStateCodeCompleted:
		response.WorkStatus = AgentWorkStatusCompleted
		response.AssignmentState = AgentAssignmentStateCompleted
		response.CommentKind = AgentTaskCommentTaskCompleted
		ensureAssignmentResponseMessage(response, "agent work completed")
	case AssignmentStateCodeFailed:
		response.WorkStatus = AgentWorkStatusFailed
		response.AssignmentState = AgentAssignmentStateFailed
		response.CommentKind = AgentTaskCommentTaskFailed
		ensureAssignmentResponseMessage(response, "agent work failed")
	default:
		ensureAssignmentResponseMessage(response, "agent assignment state updated")
	}
}

func ensureAssignmentResponseMessage(response *AIAgentTaskActionResponse, fallback string) {
	if response.Message == "" {
		response.Message = fallback
	}
}
