package riidoaiserver

func (s *DevelopmentAIAgentClientStore) markTaskAgentAssignmentThreadStoppedLocked(taskID, agentID, assignmentID string, kind AgentTaskCommentKind, message string) {
	response := AIAgentTaskActionResponse{
		WorkStatus:      AgentWorkStatusIdle,
		AssignmentState: AgentAssignmentStateStopped,
		CommentKind:     kind,
		Message:         message,
	}
	s.markTaskAgentAssignmentThreadStopProjectionLocked(taskID, agentID, assignmentID, response, true)
}
