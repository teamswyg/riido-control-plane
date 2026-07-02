package riidoaiserver

func taskThreadSupersedesQueuedEvent(thread AIAgentTaskThreadRecord) bool {
	if thread.WorkStatus == AgentWorkStatusRunning ||
		thread.AssignmentState == AgentAssignmentStateRunning ||
		len(thread.Lines) > 0 {
		return true
	}
	return thread.CommentKind != "" &&
		thread.CommentKind != AgentTaskCommentQueuedByBusyAgent
}
