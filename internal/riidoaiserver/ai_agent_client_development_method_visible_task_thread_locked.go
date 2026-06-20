package riidoaiserver

func (s *DevelopmentAIAgentClientStore) visibleTaskThreadLocked(principal AuthorizationResult, taskID, threadID string) (AIAgentTaskThreadRecord, bool) {
	for _, thread := range s.taskThreads[taskID] {
		if thread.ThreadID != threadID {
			continue
		}
		agent, ok := s.agents[thread.AgentID]
		if !ok || !s.aiAgentVisibleTo(principal, agent) {
			return AIAgentTaskThreadRecord{}, false
		}
		return copyTaskThread(thread), true
	}
	return AIAgentTaskThreadRecord{}, false
}
