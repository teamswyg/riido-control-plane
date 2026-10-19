package riidoaiserver

import (
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
