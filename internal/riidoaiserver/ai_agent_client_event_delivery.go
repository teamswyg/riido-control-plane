package riidoaiserver

import (
	"context"
	"net/http"
)

type clientEventDeliveryFilter func(context.Context, ClientStreamEvent) (ClientStreamEvent, bool)

func filterAIAgentClientDelivery(ctx context.Context, event ClientStreamEvent, filters []clientEventDeliveryFilter) (ClientStreamEvent, bool) {
	if len(filters) > 0 {
		return filters[0](ctx, event)
	}
	return event, event.EventType != AgentClientEventContextRead && event.EventType != AgentClientEventTaskNotification
}

func (s Server) clientEventForDelivery(r *http.Request, principal AuthorizationResult, event ClientStreamEvent) (ClientStreamEvent, bool) {
	if event.EventType == AgentClientEventContextRead {
		return s.clientContextReadForDelivery(r, principal, event)
	}
	return event, s.clientNotificationAuthorized(r, principal, event)
}

func writeAIAgentClientStreamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

func flushAIAgentClientStream(w http.ResponseWriter) {
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
