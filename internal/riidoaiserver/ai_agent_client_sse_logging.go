package riidoaiserver

import (
	"log"
	"net/http"
	"strings"
)

func logAIAgentClientStreamOpen(r *http.Request, replayEvents int) {
	log.Printf(
		"riido_ai_agent_sse event=stream_open path=%q replay_cursor_present=%t replay_events=%d",
		r.URL.Path, strings.TrimSpace(r.Header.Get("Last-Event-ID")) != "", replayEvents,
	)
}

func logAIAgentClientStreamClose(r *http.Request, reason string, err error) {
	errorText := ""
	if err != nil {
		errorText = err.Error()
	}
	log.Printf(
		"riido_ai_agent_sse event=stream_closed path=%q reason=%q error=%q",
		r.URL.Path, reason, errorText,
	)
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
