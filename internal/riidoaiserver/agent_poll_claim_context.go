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
		if r.Header.Get(executionCoverageHeader) != "" || strings.HasPrefix(req.DeviceID, executionCoveragePrefix) {
			writeError(w, http.StatusForbidden, "execution coverage requires device authentication")
			return nil, false
		}
		return r.Context(), true
	}
	if deviceID != strings.TrimSpace(req.DeviceID) {
		writeError(w, http.StatusForbidden, "poll device does not match authenticated device")
		return nil, false
	}
	ctx := context.WithValue(r.Context(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{deviceID: deviceID, ownerPrincipalID: auth.PrincipalID})
	if auth.executionCoverageGeneration != "" || auth.credentialGeneration != "" || strings.HasPrefix(deviceID, executionCoveragePrefix) || r.Header.Get(executionCoverageHeader) != "" {
		generation, valid := singleExecutionHeader(r, executionCoverageHeader, 512)
		if !valid || generation != auth.executionCoverageGeneration || auth.credentialGeneration == "" {
			writeError(w, http.StatusForbidden, "execution coverage generation does not match authenticated device")
			return nil, false
		}
		ctx = context.WithValue(ctx, executionCoverageAuthKey{}, executionCoverageAuth{generation, auth.credentialGeneration})
	}
	return ctx, true
}
