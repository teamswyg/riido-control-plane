package riidoaiserver

func (s *DevelopmentAIAgentClientStore) onboardingSeedThreadLocked(taskID, assignmentID string) (AIAgentTaskThreadRecord, bool) {
	for _, thread := range s.taskThreads[taskID] {
		if thread.AssignmentID == assignmentID {
			return thread, true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
