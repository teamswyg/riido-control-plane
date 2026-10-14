package riidoaiserver

func applyToolApprovalWithoutPendingActionResponse(response *AIAgentTaskActionResponse) {
	response.WorkStatus = AgentWorkStatusFailed
	response.AssignmentState = AgentAssignmentStateFailed
	response.CommentKind = AgentTaskCommentTaskFailed
	response.Message = clientMessageToolApprovalUnavailable
	response.ResultMessage = clientMessageToolApprovalUnavailable
}
