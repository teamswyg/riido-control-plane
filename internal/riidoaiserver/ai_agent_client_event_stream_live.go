package riidoaiserver

import (
	"io"
	"net/http"
	"time"
)

func streamLiveAIAgentClientEvents(w http.ResponseWriter, r *http.Request, live <-chan ClientStreamEvent, filters ...clientEventDeliveryFilter) {
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()
	drain := time.NewTimer(DefaultAIAgentTaskContextClientTimeout)
	drain.Stop()
	defer drain.Stop()
	queue := contextReadDeliveryQueue{ctx: r.Context(), filters: filters, results: make(chan contextReadDeliveryResult, 1)}
	defer queue.stop()
	closed := false
	for {
		select {
		case event, ok := <-live:
			if !ok {
				live = nil
				closed = true
				drain.Reset(DefaultAIAgentTaskContextClientTimeout)
				break
			}
			if event.EventType == AgentClientEventContextRead {
				queue.enqueue(event)
				continue
			}
			event, allowed := filterAIAgentClientDelivery(r.Context(), event, filters)
			if allowed {
				queue.supersede(event.Seq)
				if !writeLiveAIAgentClientEvent(w, r, event) {
					return
				}
			}
		case result := <-queue.results:
			if event, allowed := queue.complete(result); allowed && !writeLiveAIAgentClientEvent(w, r, event) {
				return
			}
		case <-keepalive.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				logAIAgentClientStreamClose(r, "keepalive_write_error", err)
				return
			}
			flushAIAgentClientStream(w)
		case <-drain.C:
			return
		case <-r.Context().Done():
			logAIAgentClientStreamClose(r, "request_context_done", r.Context().Err())
			return
		}
		if closed && !queue.active {
			logAIAgentClientStreamClose(r, "subscriber_closed", nil)
			return
		}
	}
}

func writeLiveAIAgentClientEvent(w http.ResponseWriter, r *http.Request, event ClientStreamEvent) bool {
	if err := writeAIAgentClientSSE(w, event); err != nil {
		logAIAgentClientStreamClose(r, "event_write_error", err)
		return false
	}
	logAIAgentClientTerminalDelivery("live_written", event, aiAgentWorkspaceIDFromRequest(r), r.Context())
	flushAIAgentClientStream(w)
	return true
}
