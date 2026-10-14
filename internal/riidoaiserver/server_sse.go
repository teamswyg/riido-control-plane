package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (s Server) streamTaskEvents(w http.ResponseWriter, r *http.Request, taskID string) {
	history, events, cancel, err := s.assignment.SubscribeTask(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)
	for _, event := range history {
		if err := writeSSE(w, event); err != nil {
			return
		}
	}
	if flusher != nil {
		flusher.Flush()
	}
	if r.URL.Query().Get("replay") == "1" {
		return
	}
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeSSE(w, event); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		case <-keepalive.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

func writeSSE(w http.ResponseWriter, event TaskEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.Seq, event.Type, data)
	return err
}
