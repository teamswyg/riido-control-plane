package riidoaiserver

func eventThreadRecord(taskID, threadID, assignmentID string, state AgentAssignmentState) AIAgentTaskThreadRecord {
	return AIAgentTaskThreadRecord{
		ThreadID:        threadID,
		TaskID:          taskID,
		AssignmentID:    assignmentID,
		AgentID:         "agent-owned-codex",
		RunID:           "run-" + assignmentID,
		AssignmentState: state,
		WorkStatus:      AgentWorkStatusRunning,
	}
}
