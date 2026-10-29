package riidoaiserver

const agentTaskNotificationStarted = "started"

// startedNotificationKindLocked sends the started push once per assignment,
// when the run starts working. runAlreadyStarted keeps it once even after the
// first push leaves the event replay buffer. Only runs started from a thread message or a
// monolith comment qualify, because those are the runs that receive 1f440.
func (s *DevelopmentAIAgentClientStore) startedNotificationKindLocked(event TaskEvent, response AIAgentTaskActionResponse) string {
	if event.Type != EventAssignmentRunning || response.ConfirmationReason != "" {
		return ""
	}
	for _, thread := range s.taskThreads[event.TaskID] {
		if thread.ThreadID != response.ThreadID {
			continue
		}
		if thread.SourceCommentID != "" || taskThreadStatusReactionTarget(thread, s.taskThreadMessages[thread.ThreadID]) != "" {
			return agentTaskNotificationStarted
		}
		return ""
	}
	return ""
}

// runAlreadyStarted reports that the stored thread already ran this
// assignment, such as a run that resumes after the user answered. It reads the
// persisted thread, not the bounded event buffer.
func runAlreadyStarted(hadThread bool, previous AIAgentTaskThreadRecord, assignmentID string) bool {
	if !hadThread || previous.AssignmentID != assignmentID {
		return false
	}
	return previous.AssignmentState == AgentAssignmentStateRunning || previous.AssignmentState == AgentAssignmentStateWaiting
}
