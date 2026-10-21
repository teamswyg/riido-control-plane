package riidoaiserver

import (
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
		logRequestFailure(r, "runtime_snapshot_decode", err)
		writeError(w, http.StatusBadRequest, "invalid runtime snapshot JSON")
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
		logRequestFailure(r, "runtime_snapshot_sync", err)
		status, message := runtimeSnapshotErrorResponse(err)
		writeError(w, status, message)
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}
