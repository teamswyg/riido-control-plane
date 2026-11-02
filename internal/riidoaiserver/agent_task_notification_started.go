package riidoaiserver

// The event buffer keeps only recent events, so the started push is recorded
// on the persisted thread to stay at most once per assignment.
func (s *DevelopmentAIAgentClientStore) startedNotificationSentLocked(taskID, assignmentID string) bool {
	if assignmentID == "" {
		return false
	}
	for _, thread := range s.taskThreads[taskID] {
		if thread.AssignmentID == assignmentID && thread.StartedNotificationSent {
			return true
		}
	}
	return false
}

func (s *DevelopmentAIAgentClientStore) markStartedNotificationSentLocked(taskID, assignmentID string) {
	if assignmentID == "" {
		return
	}
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].AssignmentID == assignmentID {
			threads[i].StartedNotificationSent = true
		}
	}
}
