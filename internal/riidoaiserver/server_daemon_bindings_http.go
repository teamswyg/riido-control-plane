package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
)

func (s Server) handleDaemonAgentBindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if s.daemonRuntime == nil {
		writeError(w, http.StatusServiceUnavailable, "daemon runtime store is not configured")
		return
	}
	deviceID := strings.TrimSpace(r.Header.Get(deviceIDHeader))
	if deviceID == "" {
		writeUnauthorized(w)
		return
	}
	principal, ok := s.authorizeRequest(w, r, AuthorizationRequest{
		Resource: AuthorizationResourceAgent,
		Action:   AuthorizationActionRead,
		DeviceID: deviceID,
	})
	if !ok {
		return
	}
	response, err := s.daemonRuntime.ListAIAgentDaemonAgentBindings(r.Context(), principal, deviceID)
	if err != nil {
		if errors.Is(err, ErrAuthorizationForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}
