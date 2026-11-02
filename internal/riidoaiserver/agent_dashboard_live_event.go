package riidoaiserver

// Never publish the ordinary queued payload to other surfaces. The dashboard
// consumes only the assignment identity and state from a queue transition.
func dashboardLiveEvent(s *DevelopmentAIAgentClientStore, subscriber aiAgentClientSubscriber, event ClientStreamEvent) (ClientStreamEvent, bool) {
	id, _ := eventAgentID(event.Payload)
	if id != subscriber.dashboardAgentID || !clientEventVisibleToLocked(s, subscriber.principal, event) {
		return ClientStreamEvent{}, false
	}
	switch v := event.Payload.(type) {
	case AgentWorkStatusChangedEvent:
		event.Payload = AgentWorkStatusChangedEvent{EventType: event.EventType, AgentID: id, TaskID: v.TaskID, ThreadID: v.ThreadID, AssignmentID: v.AssignmentID, RunID: v.RunID, AssignmentState: v.AssignmentState, WorkStatus: v.WorkStatus, CommentKind: v.CommentKind, ConfirmationReason: v.ConfirmationReason}
	case AgentThreadProgressEvent:
		v.Lines = copyClientVisibleProgressLines(v.Lines)
		event.Payload = v
	default:
		return ClientStreamEvent{}, false
	}
	return event, true
}

func dashboardEventTaskID(event ClientStreamEvent) string {
	switch v := event.Payload.(type) {
	case AgentWorkStatusChangedEvent:
		return v.TaskID
	case AgentThreadProgressEvent:
		return v.TaskID
	}
	return ""
}

func (s *DevelopmentAIAgentClientStore) deliverDashboardEventLocked(id int, sub aiAgentClientSubscriber, event ClientStreamEvent) {
	visible, ok := dashboardLiveEvent(s, sub, event)
	if !ok {
		return
	}
	select {
	case sub.events <- visible:
	default:
		// A slow consumer reconnects and gets a fresh, authorized snapshot.
		delete(s.subscribers, id)
		close(sub.events)
	}
}

// Restoring a shared snapshot can reveal transitions committed on another
// replica. Keep existing subscribers, and forward only events past the local cursor.
func (s *DevelopmentAIAgentClientStore) deliverRestoredDashboardEventsLocked(next int64) {
	for _, e := range s.events {
		if e.Seq < next {
			continue
		}
		for id, sub := range s.subscribers {
			if sub.dashboardAgentID != "" {
				s.deliverDashboardEventLocked(id, sub, e)
			}
		}
	}
}
