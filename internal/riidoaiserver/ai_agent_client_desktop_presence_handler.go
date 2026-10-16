package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"time"
)

func (s Server) handleAIAgentClientDesktopPresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if aiAgentWorkspaceIDFromRequest(r) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	reader, ok := s.aiAgent.(AIAgentDesktopPresenceReader)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "desktop presence is not configured")
		return
	}
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead})
	if !ok {
		return
	}
	lastSeenAt, err := reader.DesktopPresenceLastSeenAt(r.Context(), principal)
	if err != nil {
		writeDesktopPresenceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newAIAgentDesktopPresenceResponse(lastSeenAt, time.Now()))
}

// This read exposes no dependency diagnostics or stored values to callers.
func writeDesktopPresenceError(w http.ResponseWriter, err error) {
	status, message := http.StatusServiceUnavailable, "desktop presence unavailable"
	switch {
	case errors.Is(err, ErrAuthorizationForbidden):
		status, message = http.StatusForbidden, "forbidden"
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusGatewayTimeout
	case errors.Is(err, context.Canceled):
		status = http.StatusRequestTimeout
	}
	writeError(w, status, message)
}
