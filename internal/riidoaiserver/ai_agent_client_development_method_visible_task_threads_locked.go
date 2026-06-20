package riidoaiserver

func (s *DevelopmentAIAgentClientStore) visibleTaskThreadsLocked(principal AuthorizationResult, taskID string) []AIAgentTaskThreadRecord {
	source := s.taskThreads[taskID]
	out := make([]AIAgentTaskThreadRecord, 0, len(source))
	for _, thread := range source {
		agent, ok := s.agents[thread.AgentID]
		if !ok || !s.aiAgentVisibleTo(principal, agent) {
			continue
		}
		out = append(out, copyTaskThread(thread))
	}
	return out
}
