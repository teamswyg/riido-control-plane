package riidoaiserver

import (
	"io"
	"net/http"
	"time"
)

func serveAIAgentClientEventStream(w http.ResponseWriter, r *http.Request, events []ClientStreamEvent, live <-chan ClientStreamEvent) {
	writeAIAgentClientStreamHeaders(w)
	for _, event := range events {
		if err := writeAIAgentClientSSE(w, event); err != nil {
			return
		}
	}
	flushAIAgentClientStream(w)
	if r.URL.Query().Get("replay") == "1" {
		return
	}
	streamLiveAIAgentClientEvents(w, r, live)
}

func writeAIAgentClientStreamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

func streamLiveAIAgentClientEvents(w http.ResponseWriter, r *http.Request, live <-chan ClientStreamEvent) {
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()
	for {
		select {
		case event, ok := <-live:
			if !ok {
				return
			}
			if err := writeAIAgentClientSSE(w, event); err != nil {
				return
			}
			flushAIAgentClientStream(w)
		case <-keepalive.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
			flushAIAgentClientStream(w)
		case <-r.Context().Done():
			return
		}
	}
}

func flushAIAgentClientStream(w http.ResponseWriter) {
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
