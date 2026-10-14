package riidoaiserver

func applyAssignmentStatusToAgent(agent AgentClientRecord, response AIAgentTaskActionResponse) AgentClientRecord {
	if !taskThreadHasActiveStream(AIAgentTaskThreadRecord{AssignmentState: response.AssignmentState}) &&
		agent.AssignedTaskCount > 0 {
		agent.AssignedTaskCount--
	}
	agent.WorkStatus = response.WorkStatus
	agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
	if taskThreadHasActiveStream(AIAgentTaskThreadRecord{AssignmentState: response.AssignmentState}) {
		agent.Editability = AgentEditabilityBlockedAssignedTasks
		if agent.AssignedTaskCount == 0 {
			agent.AssignedTaskCount = 1
		}
	}
	return agent
}
