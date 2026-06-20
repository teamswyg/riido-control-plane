package riidoaiserver

func (s *DevelopmentAIAgentClientStore) projectAgentWorkStatusFromThreadsLocked(agent AgentClientRecord) AgentClientRecord {
	activeCount := 0
	var latest AIAgentTaskThreadRecord
	hasLatest := false
	for _, threads := range s.taskThreads {
		for _, thread := range threads {
			if thread.AgentID != agent.AgentID || !taskThreadHasActiveStream(thread) {
				continue
			}
			activeCount++
			if !hasLatest ||
				thread.StartedAt.After(latest.StartedAt) ||
				(thread.StartedAt.Equal(latest.StartedAt) && thread.ThreadID > latest.ThreadID) {
				latest = thread
				hasLatest = true
			}
		}
	}
	agent.AssignedTaskCount = activeCount
	agent.Editability = editabilityForAssignedTasks(activeCount)
	if activeCount == 0 {
		switch agent.WorkStatus {
		case AgentWorkStatusQueued, AgentWorkStatusRunning, AgentWorkStatusWaitingForUser:
			agent.WorkStatus = AgentWorkStatusIdle
		default:
		}
		return agent
	}
	agent.WorkStatus = projectedAgentWorkStatusFromActiveThread(latest)
	return agent
}
