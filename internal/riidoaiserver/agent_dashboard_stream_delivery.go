package riidoaiserver

import (
	"errors"
	"net/http"
	"time"
)

func (s Server) deliverDashboardEvents(w http.ResponseWriter, r *http.Request, live <-chan ClientStreamEvent, workspace, agentID string) {
	ctx := r.Context()
	token, _ := requestToken(r)
	keepalive := time.NewTicker(sseKeepaliveInterval)
	defer keepalive.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepalive.C:
			if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
				return
			}
			flushAIAgentClientStream(w)
		case event, open := <-live:
			if !open {
				return
			}
			id, _ := eventAgentID(event.Payload)
			if id != agentID {
				continue
			}
			taskID := dashboardEventTaskID(event)
			if taskID == "" {
				continue
			}
			// A shared agent can execute tasks in teams the viewer cannot read.
			_, err := s.dashboardTaskForViewer(ctx, token, workspace, taskID)
			if errors.Is(err, ErrAuthorizationForbidden) {
				continue
			}
			if err != nil {
				return
			}
			if err := writeAIAgentClientSSE(w, event); err != nil {
				return
			}
			flushAIAgentClientStream(w)
		}
	}
}
