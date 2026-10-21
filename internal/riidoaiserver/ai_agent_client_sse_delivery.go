package riidoaiserver

import (
	"context"
	"log"
)

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
	if !clientStreamEventIsTerminalProgress(event) && event.EventType != AgentClientEventTaskNotification {
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
		"riido_ai_agent_sse event=terminal_delivery_compensated drops=%d assignment_ref=%q thread_ref=%q run_ref=%q state=%q",
		drops, providerHealthDeviceRef(progress.AssignmentID), providerHealthDeviceRef(progress.ThreadID), providerHealthDeviceRef(progress.RunID), boundedTerminalState(progress.AssignmentState),
	)
}

func logAIAgentClientTerminalDelivery(phase string, event ClientStreamEvent, workspaceID string, contexts ...context.Context) {
	if !clientStreamEventIsTerminalProgress(event) {
		return
	}
	progress, _ := event.Payload.(AgentThreadProgressEvent)
	ctx := context.Background()
	if len(contexts) > 0 {
		ctx = contexts[0]
	}
	ctx, span := startSSEObservation(ctx, "terminal_delivery", StringTraceAttribute("riido.sse.delivery_phase", boundedTerminalPhase(phase)), StringTraceAttribute("riido.assignment.state", boundedTerminalState(progress.AssignmentState)))
	defer span.End()
	traceID, spanID := traceCorrelationFromContext(ctx)
	log.Printf(
		"riido_ai_agent_sse event=terminal_delivery phase=%q workspace_ref=%q assignment_ref=%q thread_ref=%q run_ref=%q state=%q trace_id=%q span_id=%q",
		boundedTerminalPhase(phase), providerHealthDeviceRef(workspaceID), providerHealthDeviceRef(progress.AssignmentID), providerHealthDeviceRef(progress.ThreadID), providerHealthDeviceRef(progress.RunID), boundedTerminalState(progress.AssignmentState), traceID, spanID,
	)
}

func boundedTerminalPhase(phase string) string {
	switch phase {
	case "fanout_enqueued", "fanout_enqueued_after_overflow", "replay_written", "live_written":
		return phase
	default:
		return "unknown"
	}
}

func boundedTerminalState(state AgentAssignmentState) string {
	switch state {
	case AgentAssignmentStateCompleted, AgentAssignmentStateFailed, AgentAssignmentStateStopped:
		return string(state)
	default:
		return "unknown"
	}
}
