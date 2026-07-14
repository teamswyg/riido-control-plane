package riidoaiserver

import (
	"errors"
	"net/http"
	"strings"
)

func (s Server) authorizeDeviceCredential(w http.ResponseWriter, r *http.Request, req AuthorizationRequest) (AuthorizationResult, bool) {
	deviceID := strings.TrimSpace(r.Header.Get(deviceIDHeader))
	deviceSecret := strings.TrimSpace(r.Header.Get(deviceSecretHeader))
	if deviceID == "" && deviceSecret == "" {
		return AuthorizationResult{}, false
	}
	if deviceID == "" || deviceSecret == "" {
		writeUnauthorized(w)
		return AuthorizationResult{}, true
	}
	if s.devices == nil {
		writeError(w, http.StatusServiceUnavailable, "device credential store is not configured")
		return AuthorizationResult{}, true
	}
	if req.DeviceID == "" {
		req.DeviceID = deviceID
	}
	result, err := s.devices.AuthorizeDeviceCredential(r.Context(), deviceID, deviceSecret, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthorizationForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAuthorizationUnauthenticated):
			writeUnauthorized(w)
		case errors.Is(err, ErrAuthorizationInvalidCredential):
			writeUnauthorized(w)
		default:
			writeError(w, http.StatusServiceUnavailable, err.Error())
		}
		return AuthorizationResult{}, true
	}
	return result, true
}
