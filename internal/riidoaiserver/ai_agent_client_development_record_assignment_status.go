package riidoaiserver

func (s *DevelopmentAIAgentClientStore) recordAssignmentStatusEventLocked(
	input assignmentEventInput,
	agent AgentClientRecord,
	hadThread bool,
	previousThread AIAgentTaskThreadRecord,
	thread AIAgentTaskThreadRecord,
) {
	message := assignmentEventVisibleThreadMessage(input.State, input.Type, input.Message, previousThread.Message)
	response := assignmentEventActionResponse(thread, input.State, message, input.Metadata)
	response.AssignmentID = input.AssignmentID
	s.upsertTaskThreadFromActionLocked(response, "")
	shouldFanoutStatus := shouldFanoutAgentTaskActionEvent(hadThread, previousThread, response)
	agent = applyAssignmentStatusToAgent(agent, response)
	s.agents[input.AgentID] = agent
	if shouldFanoutStatus {
		s.appendAgentTaskActionEvent(response)
	}
}
