package riidoaiserver

func (s *DevelopmentAIAgentClientStore) activeTaskThreadStreamTargetsLocked(
	principal AuthorizationResult,
	taskID string,
) []AIAgentTaskThreadStreamTarget {
	source := s.taskThreads[taskID]
	targets := make([]AIAgentTaskThreadStreamTarget, 0, activeStreamTargetCapacity(source))
	for i := range source {
		s.ensureTaskThreadAgentSnapshotLocked(&source[i], source[i].StartedAt)
		if !taskThreadHasActiveStream(source[i]) || !s.taskThreadVisibleTo(principal, source[i]) {
			continue
		}
		targets = append(targets, AIAgentTaskThreadStreamTarget{
			AgentID:  source[i].AgentID,
			ThreadID: source[i].ThreadID,
			RunID:    source[i].RunID,
		})
	}
	s.taskThreads[taskID] = source
	return targets
}

func activeStreamTargetCapacity(threads []AIAgentTaskThreadRecord) int {
	if len(threads) == 0 || taskThreadHasActiveStream(threads[0]) {
		return len(threads)
	}
	count := 0
	for i := range threads {
		if taskThreadHasActiveStream(threads[i]) {
			count++
		}
	}
	return count
}
