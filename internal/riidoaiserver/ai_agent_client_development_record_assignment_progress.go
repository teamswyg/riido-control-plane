package riidoaiserver

func (s *DevelopmentAIAgentClientStore) recordAssignmentProgressEventLocked(
	input assignmentEventInput,
	agent AgentClientRecord,
	hadThread bool,
	previousThread AIAgentTaskThreadRecord,
	thread AIAgentTaskThreadRecord,
) error {
	if hadThread && !agentAssignmentStateAcceptsRuntimeProgress(previousThread.AssignmentState) {
		return nil
	}
	line := s.assignmentProgressLineLocked(input, thread)
	if existing, ok := s.taskThreadByIDLocked(thread.TaskID, thread.ThreadID); ok &&
		progressLineSeqSeen(existing.Lines, line.Seq) {
		return nil
	}
	event := assignmentProgressEvent(input, thread, line)
	if thread.ConfirmationReason != "" {
		event.WorkStatus = AgentWorkStatusWaitingForUser
	}
	s.appendThreadProgressLocked(event)
	s.appendClientEventLocked(event.EventType, event)
	agent = markAgentRunningFromAssignmentProgress(agent)
	if thread.ConfirmationReason != "" {
		agent.WorkStatus = AgentWorkStatusWaitingForUser
	}
	s.agents[input.AgentID] = agent
	return nil
}
