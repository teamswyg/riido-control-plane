package riidoaiserver

import "log"

func deliverAIAgentClientEvent(
	subscriber aiAgentClientSubscriber,
	event ClientStreamEvent,
) aiAgentClientSubscriber {
	select {
	case subscriber.events <- event:
		logAIAgentClientTerminalDelivery("fanout_enqueued", event, subscriber.principal.WorkspaceID)
		return subscriber
	default:
	}
	subscriber.droppedEvents++
	if !clientStreamEventIsTerminalProgress(event) {
		return subscriber
	}
	select {
	case <-subscriber.events:
	default:
	}
	subscriber.terminalCompensations++
	subscriber.events <- event
	logAIAgentClientTerminalPriority(event, subscriber.droppedEvents)
	logAIAgentClientTerminalDelivery("fanout_enqueued_after_overflow", event, subscriber.principal.WorkspaceID)
	return subscriber
}

func logAIAgentClientTerminalPriority(event ClientStreamEvent, drops int64) {
	progress, _ := event.Payload.(AgentThreadProgressEvent)
	log.Printf(
		"riido_ai_agent_sse event=terminal_delivery_compensated drops=%d assignment_id=%q thread_id=%q run_id=%q state=%q",
		drops, progress.AssignmentID, progress.ThreadID, progress.RunID, progress.AssignmentState,
	)
}

func logAIAgentClientTerminalDelivery(phase string, event ClientStreamEvent, workspaceID string) {
	if !clientStreamEventIsTerminalProgress(event) {
		return
	}
	progress, _ := event.Payload.(AgentThreadProgressEvent)
	log.Printf(
		"riido_ai_agent_sse event=terminal_delivery phase=%q workspace_id=%q assignment_id=%q thread_id=%q run_id=%q state=%q",
		phase, workspaceID, progress.AssignmentID, progress.ThreadID, progress.RunID, progress.AssignmentState,
	)
}
