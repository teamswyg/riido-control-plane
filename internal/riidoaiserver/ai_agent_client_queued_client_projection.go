package riidoaiserver

func clientVisibleQueuedActionResponse(response AIAgentTaskActionResponse) AIAgentTaskActionResponse {
	if !actionResponseIsQueuedForClient(response) {
		return response
	}
	response.WorkStatus = AgentWorkStatusIdle
	response.AssignmentState = ""
	response.CommentKind = ""
	response.Message = ""
	response.ResultMessage = ""
	return response
}

func clientVisibleQueuedTaskThread(thread AIAgentTaskThreadRecord) AIAgentTaskThreadRecord {
	if !taskThreadIsQueuedForClient(thread) {
		return thread
	}
	thread.WorkStatus = AgentWorkStatusIdle
	thread.AssignmentState = ""
	thread.CommentKind = ""
	thread.Message = ""
	thread.ResultMessage = ""
	thread.QueueDiagnostics = nil
	return thread
}

func actionResponseIsQueuedForClient(response AIAgentTaskActionResponse) bool {
	return response.CommentKind == AgentTaskCommentQueuedByBusyAgent ||
		response.WorkStatus == AgentWorkStatusQueued ||
		response.AssignmentState == AgentAssignmentStateQueued
}

func taskThreadIsQueuedForClient(thread AIAgentTaskThreadRecord) bool {
	return thread.CommentKind == AgentTaskCommentQueuedByBusyAgent ||
		thread.WorkStatus == AgentWorkStatusQueued ||
		thread.AssignmentState == AgentAssignmentStateQueued
}

func eventIsQueuedForClient(event ClientStreamEvent) bool {
	status, ok := event.Payload.(AgentWorkStatusChangedEvent)
	return ok && workStatusEventIsQueuedForClient(status)
}

func workStatusEventIsQueuedForClient(status AgentWorkStatusChangedEvent) bool {
	return status.CommentKind == AgentTaskCommentQueuedByBusyAgent ||
		status.WorkStatus == AgentWorkStatusQueued ||
		status.AssignmentState == AgentAssignmentStateQueued
}
