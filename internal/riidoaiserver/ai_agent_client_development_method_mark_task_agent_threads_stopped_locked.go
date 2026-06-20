package riidoaiserver

func (s *DevelopmentAIAgentClientStore) markTaskAgentThreadsStoppedLocked(taskID, agentID string, kind AgentTaskCommentKind, message string) {
	response := AIAgentTaskActionResponse{
		WorkStatus:      AgentWorkStatusIdle,
		AssignmentState: AgentAssignmentStateStopped,
		CommentKind:     kind,
		Message:         message,
	}
	s.markTaskAgentThreadsStopProjectionLocked(taskID, agentID, response, true)
}
