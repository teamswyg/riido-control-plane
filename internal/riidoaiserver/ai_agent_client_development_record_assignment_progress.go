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
	s.appendThreadProgressLocked(event)
	s.appendClientEventLocked(event.EventType, event)
	s.agents[input.AgentID] = markAgentRunningFromAssignmentProgress(agent)
	return nil
}
