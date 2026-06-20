package riidoaiserver

func markAgentRunningFromAssignmentProgress(agent AgentClientRecord) AgentClientRecord {
	agent.WorkStatus = AgentWorkStatusRunning
	agent.Editability = AgentEditabilityBlockedAssignedTasks
	if agent.AssignedTaskCount == 0 {
		agent.AssignedTaskCount = 1
	}
	return agent
}
