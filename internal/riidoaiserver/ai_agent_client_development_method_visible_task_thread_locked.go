package riidoaiserver

func (s *DevelopmentAIAgentClientStore) visibleTaskThreadLocked(principal AuthorizationResult, taskID, threadID string) (AIAgentTaskThreadRecord, bool) {
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].ThreadID != threadID {
			continue
		}
		s.ensureTaskThreadAgentSnapshotLocked(&threads[i], threads[i].StartedAt)
		if !s.taskThreadVisibleTo(principal, threads[i]) {
			return AIAgentTaskThreadRecord{}, false
		}
		s.taskThreads[taskID] = threads
		return copyTaskThread(threads[i]), true
	}
	return AIAgentTaskThreadRecord{}, false
}
