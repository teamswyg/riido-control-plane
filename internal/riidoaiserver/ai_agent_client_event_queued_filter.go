package riidoaiserver

func withoutSupersededQueuedClientEvents(
	s *DevelopmentAIAgentClientStore,
	events []ClientStreamEvent,
) []ClientStreamEvent {
	out := events[:0]
	for i, event := range events {
		if queuedClientEventIsSupersededLocked(s, event, events[i+1:]) {
			continue
		}
		out = append(out, event)
	}
	return out
}

func queuedClientEventIsSupersededLocked(
	s *DevelopmentAIAgentClientStore,
	event ClientStreamEvent,
	later []ClientStreamEvent,
) bool {
	status, ok := event.Payload.(AgentWorkStatusChangedEvent)
	if !ok || status.CommentKind != AgentTaskCommentQueuedByBusyAgent {
		return false
	}
	thread, ok := s.rawTaskThreadByIDLocked(status.TaskID, status.ThreadID)
	if !ok {
		return false
	}
	if taskThreadSupersedesQueuedEvent(thread) {
		return true
	}
	return laterEventsSupersedeQueuedLocked(s, taskThreadConversationID(thread), later)
}

func laterEventsSupersedeQueuedLocked(
	s *DevelopmentAIAgentClientStore,
	conversationID string,
	events []ClientStreamEvent,
) bool {
	for _, event := range events {
		if !clientEventSupersedesQueued(event) {
			continue
		}
		if eventConversationIDLocked(s, event) == conversationID {
			return true
		}
	}
	return false
}

func eventConversationIDLocked(
	s *DevelopmentAIAgentClientStore,
	event ClientStreamEvent,
) string {
	taskID, threadID, ok := eventTaskThreadRef(event.Payload)
	if !ok {
		return ""
	}
	thread, ok := s.rawTaskThreadByIDLocked(taskID, threadID)
	if !ok {
		return ""
	}
	return taskThreadConversationID(thread)
}

func clientEventSupersedesQueued(event ClientStreamEvent) bool {
	if _, ok := event.Payload.(AgentThreadProgressEvent); ok {
		return true
	}
	status, ok := event.Payload.(AgentWorkStatusChangedEvent)
	return ok && status.CommentKind != AgentTaskCommentQueuedByBusyAgent
}
