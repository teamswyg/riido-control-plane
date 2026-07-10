package riidoaiserver

import "log"

func deliverAIAgentClientEvent(
	subscriber aiAgentClientSubscriber,
	event ClientStreamEvent,
) aiAgentClientSubscriber {
	select {
	case subscriber.events <- event:
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
	return subscriber
}

func logAIAgentClientTerminalPriority(event ClientStreamEvent, drops int64) {
	progress, _ := event.Payload.(AgentThreadProgressEvent)
	log.Printf(
		"riido_ai_agent_sse event=terminal_delivery_compensated drops=%d assignment_id=%q thread_id=%q run_id=%q state=%q",
		drops, progress.AssignmentID, progress.ThreadID, progress.RunID, progress.AssignmentState,
	)
}
