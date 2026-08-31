package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
)

func (s Server) handleDaemonRuntimeSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
		Action:   AuthorizationActionProviderStatusWrite,
		DeviceID: deviceID,
	})
	if !ok {
		return
	}
	var req DeviceRuntimeSnapshotSyncRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.DeviceID == "" {
		req.DeviceID = deviceID
	}
	if req.DeviceID != deviceID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	response, err := s.daemonRuntime.SyncAIAgentDaemonRuntimeSnapshot(r.Context(), principal, req)
	if err != nil {
		if errors.Is(err, ErrAuthorizationForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	traceProviderHealth(r.Context(), response.Device.Runtimes)
	writeJSON(w, http.StatusAccepted, response)
}
