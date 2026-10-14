package riidoaiserver

func (s *DevelopmentAIAgentClientStore) visibleTaskThreadsLocked(principal AuthorizationResult, taskID string) []AIAgentTaskThreadRecord {
	source := s.taskThreads[taskID]
	out := make([]AIAgentTaskThreadRecord, 0, len(source))
	for i := range source {
		s.ensureTaskThreadAgentSnapshotLocked(&source[i], source[i].StartedAt)
		if !s.taskThreadVisibleTo(principal, source[i]) {
			continue
		}
		out = append(out, copyTaskThread(source[i]))
	}
	s.taskThreads[taskID] = source
	return out
}
