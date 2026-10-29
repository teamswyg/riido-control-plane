package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"time"
)

func (s Server) handleExecutionTermination(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if recover() != nil {
			writeTerminationError(w, errTerminationUnavailable)
		}
	}()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	identity, err := s.authorizeExecutionTermination(r)
	if err != nil {
		writeTerminationError(w, err)
		return
	}
	req, err := decodeExecutionTermination(w, r)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "execution termination body is too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid execution termination observation")
		}
		return
	}
	recorder, ok := s.assignment.(executionTerminationRecorder)
	if !ok {
		writeTerminationError(w, errTerminationUnavailable)
		return
	}
	receipt, err := recorder.recordExecutionTermination(ctx, identity, req)
	if err != nil {
		writeTerminationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}

func writeTerminationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrAuthorizationUnauthenticated), errors.Is(err, ErrAuthorizationInvalidCredential):
		writeError(w, http.StatusUnauthorized, "device and execution proof required")
	case errors.Is(err, errTerminationForbidden), errors.Is(err, ErrAuthorizationForbidden):
		writeError(w, http.StatusForbidden, "execution termination proof is not authorized")
	case errors.Is(err, errTerminationConflict):
		writeError(w, http.StatusConflict, "execution termination observation conflicts")
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "execution termination timed out; retry the same observation")
	case errors.Is(err, context.Canceled):
		writeError(w, http.StatusRequestTimeout, "execution termination cancelled; retry the same observation")
	default:
		writeError(w, http.StatusServiceUnavailable, "execution termination unavailable; retry the same observation")
	}
}
