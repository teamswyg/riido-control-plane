package riidoaiserver

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
)

func logAIAgentClientStreamOpen(r *http.Request, replayEvents int) {
	traceID, spanID := traceCorrelationFromContext(r.Context())
	log.Printf(
		"riido_ai_agent_sse event=stream_open route=%q replay_cursor_present=%t replay_events=%d trace_id=%q span_id=%q",
		sseLogRoute(r), strings.TrimSpace(r.Header.Get("Last-Event-ID")) != "", replayEvents, traceID, spanID,
	)
}

func logAIAgentClientStreamClose(r *http.Request, reason string, err error) {
	traceID, spanID := traceCorrelationFromContext(r.Context())
	log.Printf(
		"riido_ai_agent_sse event=stream_closed route=%q reason=%q error_type=%q trace_id=%q span_id=%q",
		sseLogRoute(r), boundedSSECloseReason(reason), sseCloseErrorType(reason, err), traceID, spanID,
	)
}

func sseLogRoute(r *http.Request) string {
	if route, ok := r.Context().Value(requestLogRouteContextKey{}).(string); ok && route != "" {
		return route
	}
	if route := traceHTTPRoute(r.Method, r.URL.Path); route != "" {
		return route
	}
	return "unknown"
}

func boundedSSECloseReason(reason string) string {
	switch reason {
	case "replay_write_error", "replay_complete", "subscriber_closed", "event_write_error", "keepalive_write_error", "request_context_done":
		return reason
	default:
		return "unknown"
	}
}

func sseCloseErrorType(reason string, err error) string {
	switch {
	case err == nil:
		return "none"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	}
	switch boundedSSECloseReason(reason) {
	case "replay_write_error", "event_write_error", "keepalive_write_error":
		return "write_error"
	default:
		return "unknown"
	}
}

func logAIAgentClientSubscriberDeliverySummary(subscriber aiAgentClientSubscriber) {
	if subscriber.droppedEvents == 0 {
		return
	}
	log.Printf(
		"riido_ai_agent_sse event=fanout_overflow_summary drops=%d terminal_compensations=%d",
		subscriber.droppedEvents, subscriber.terminalCompensations,
	)
}
