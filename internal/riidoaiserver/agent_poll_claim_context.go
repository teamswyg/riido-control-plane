package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
)

// Called only after authorizeRequest succeeds. Bearer callers never gain device provenance.
func authenticatedPollContext(w http.ResponseWriter, r *http.Request, req PollRequest, auth AuthorizationResult) (context.Context, bool) {
	deviceID := strings.TrimSpace(r.Header.Get(deviceIDHeader))
	if deviceID == "" {
		return r.Context(), true
	}
	if deviceID != strings.TrimSpace(req.DeviceID) {
		writeError(w, http.StatusForbidden, "poll device does not match authenticated device")
		return nil, false
	}
	return context.WithValue(r.Context(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{deviceID, auth.PrincipalID}), true
}
