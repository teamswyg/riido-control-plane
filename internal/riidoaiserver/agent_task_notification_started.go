package riidoaiserver

const agentTaskNotificationStarted = "started"

// startedNotificationKindLocked sends the started push once per assignment,
// when the run starts working. Only runs started from a thread message or a
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
