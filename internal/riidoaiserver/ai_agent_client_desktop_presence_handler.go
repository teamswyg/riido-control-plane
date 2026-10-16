package riidoaiserver

import (
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
		writeAIAgentClientError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newAIAgentDesktopPresenceResponse(lastSeenAt, time.Now()))
}
