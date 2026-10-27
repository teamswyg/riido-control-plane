package riidoaiserver

import (
	"context"
	"net/http"
)

func serveAIAgentClientEventStream(w http.ResponseWriter, r *http.Request, events []ClientStreamEvent, live <-chan ClientStreamEvent, filters ...clientEventDeliveryFilter) {
	writeAIAgentClientStreamHeaders(w)
	logAIAgentClientStreamOpen(r, len(events))
	// Read hints share one replay budget; progress/terminal events continue.
	ctx, cancel := context.WithTimeout(r.Context(), DefaultAIAgentTaskContextClientTimeout)
	defer cancel()
	for _, event := range events {
		if event.EventType == AgentClientEventTaskNotification {
			continue
		}
		event, allowed := filterAIAgentClientDelivery(ctx, event, filters)
		if !allowed {
			continue
		}
		if err := writeAIAgentClientSSE(w, event); err != nil {
			logAIAgentClientStreamClose(r, "replay_write_error", err)
			return
		}
		logAIAgentClientTerminalDelivery("replay_written", event, aiAgentWorkspaceIDFromRequest(r), r.Context())
	}
	flushAIAgentClientStream(w)
	if r.URL.Query().Get("replay") == "1" {
		logAIAgentClientStreamClose(r, "replay_complete", nil)
		return
	}
	cancel()
	streamLiveAIAgentClientEvents(w, r, live, filters...)
}
