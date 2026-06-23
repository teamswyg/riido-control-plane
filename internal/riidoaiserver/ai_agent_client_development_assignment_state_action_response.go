package riidoaiserver

func applyAssignmentStateActionResponse(response *AIAgentTaskActionResponse, state AssignmentState) {
	switch state.Code() {
	case AssignmentStateCodeQueued, AssignmentStateCodeLeased:
		response.WorkStatus = AgentWorkStatusQueued
		response.AssignmentState = AgentAssignmentStateQueued
		response.CommentKind = AgentTaskCommentQueuedByBusyAgent
		ensureAssignmentResponseMessage(response, clientMessageAgentBusyQueued)
	case AssignmentStateCodeReady:
		response.WorkStatus = AgentWorkStatusRunning
		response.AssignmentState = AgentAssignmentStateRunning
		response.CommentKind = AgentTaskCommentAssignmentStarted
		ensureAssignmentResponseMessage(response, clientMessageTaskRunning)
	case AssignmentStateCodeRunning:
		response.WorkStatus = AgentWorkStatusRunning
		response.AssignmentState = AgentAssignmentStateRunning
		response.CommentKind = AgentTaskCommentRuntimeProgress
		ensureAssignmentResponseMessage(response, clientMessageTaskRunning)
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
