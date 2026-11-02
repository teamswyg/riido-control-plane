package riidoaiserver

import (
	"context"
	"net/http"
	"time"
)

type dashboardStreamScope struct{}

// Register the live subscriber before signaling ready. The client then takes
// an authorized snapshot and reconciles events that arrived during that read.
func (s Server) handleDashboardStream(w http.ResponseWriter, r *http.Request, agentID string) {
	workspace := aiAgentWorkspaceIDFromRequest(r)
	if workspace == "" {
		writeError(w, http.StatusBadRequest, "workspace required")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionStream, AgentID: agentID})
	if !ok {
		return
	}
	subscriber, ok := s.aiAgent.(AIAgentClientEventSubscriber)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "dashboard stream unavailable")
		return
	}
	ctx := context.WithValue(r.Context(), dashboardStreamScope{}, agentID)
	_, live, cancel, err := subscriber.SubscribeAIAgentClientEvents(ctx, principal)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "dashboard stream unavailable")
		return
	}
	if cancel != nil {
		defer cancel()
	}
	if live == nil {
		writeError(w, http.StatusServiceUnavailable, "dashboard live events unavailable")
		return
	}
	writeAIAgentClientStreamHeaders(w)
	if _, err := w.Write([]byte(": dashboard stream ready\n\n")); err != nil {
		return
	}
	flushAIAgentClientStream(w)
	ticker := time.NewTicker(sseKeepaliveInterval)
	defer ticker.Stop()
	s.deliverDashboardEvents(w, r, live, workspace, agentID, ticker.C)
}
