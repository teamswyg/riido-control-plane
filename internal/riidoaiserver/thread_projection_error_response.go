package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
)

func writeThreadProjectionError(w http.ResponseWriter, r *http.Request, err error) {
	logRequestFailure(r, "projection_reconcile", err)
	status, message := threadProjectionErrorResponse(err)
	writeError(w, status, message)
}

func threadProjectionErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout, "thread projection request cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "thread projection synchronization timed out"
	default:
		return http.StatusBadGateway, "thread projection synchronization unavailable"
	}
}
