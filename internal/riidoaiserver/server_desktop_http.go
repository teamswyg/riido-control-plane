package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
)

func (s Server) handleDesktopWorkspaceRoutes(w http.ResponseWriter, r *http.Request) {
	workspaceID, suffix, ok := splitNestedResourcePath(r.URL.Path, "/v2/desktop/workspaces/")
	if !ok || strings.TrimSpace(workspaceID) == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	suffix = strings.Trim(suffix, "/")
	switch {
	case suffix == "devices/enroll" && r.Method == http.MethodPost:
		s.handleDesktopDeviceEnroll(w, r, workspaceID)
	case suffix == "devices/connect" && r.Method == http.MethodPost:
		s.handleDesktopDeviceConnect(w, r, workspaceID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

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
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.MachineID) == "" {
		writeError(w, http.StatusBadRequest, "machine_id is required")
		return
	}
	device, err := s.daemonRuntime.ConnectAIAgentDevice(r.Context(), principal, req.MachineID)
	if err != nil {
		if errors.Is(err, ErrAuthorizationForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"schema_version":          SchemaVersion,
		"device_id":               device.DeviceID,
		"connected_workspace_ids": device.ConnectedWorkspaceIDs,
	})
}
