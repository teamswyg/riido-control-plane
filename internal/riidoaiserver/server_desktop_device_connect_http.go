package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

func (s Server) handleDesktopDeviceConnect(w http.ResponseWriter, r *http.Request, workspaceID string) {
	if s.daemonRuntime == nil {
		writeError(w, http.StatusServiceUnavailable, "daemon runtime store is not configured")
		return
	}
	r = requestWithAIAgentWorkspaceIDAndPath(r, workspaceID, r.URL.Path)
	principal, ok := s.authorizeAIAgentClient(w, r, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead})
	if !ok {
		return
	}
	var req struct {
		MachineID string `json:"machine_id"`
		DeviceID  string `json:"device_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.MachineID) == "" && strings.TrimSpace(req.DeviceID) == "" {
		writeError(w, http.StatusBadRequest, "machine_id is required")
		return
	}
	var device DeviceRecord
	var err error
	if strings.TrimSpace(req.DeviceID) != "" || r.Header.Get(deviceSecretHeader) != "" {
		connector, supported := s.daemonRuntime.(CredentialDeviceConnector)
		if !supported {
			writeError(w, http.StatusServiceUnavailable, "credential device connection unavailable")
			return
		}
		device, err = connector.ConnectAIAgentDeviceWithCredential(r.Context(), principal, req.DeviceID, r.Header.Get(deviceSecretHeader))
	} else {
		device, err = s.daemonRuntime.ConnectAIAgentDevice(r.Context(), principal, req.MachineID)
	}
	if err != nil {
		if errors.Is(err, ErrAuthorizationUnauthenticated) {
			writeUnauthorized(w)
			return
		}
		if errors.Is(err, ErrAuthorizationForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		status, message := http.StatusServiceUnavailable, "device connection unavailable"
		if errors.Is(err, context.DeadlineExceeded) {
			status, message = http.StatusGatewayTimeout, "device connection timed out"
		}
		writeError(w, status, message)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"schema_version":          SchemaVersion,
		"device_id":               device.DeviceID,
		"connected_workspace_ids": device.ConnectedWorkspaceIDs,
	})
}
