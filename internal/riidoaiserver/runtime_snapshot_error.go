package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
)

type runtimeSnapshotValidationError struct{ cause error }

func (e runtimeSnapshotValidationError) Error() string { return e.cause.Error() }
func (e runtimeSnapshotValidationError) Unwrap() error { return e.cause }
func invalidRuntimeSnapshot(err error) error           { return runtimeSnapshotValidationError{cause: err} }

func runtimeSnapshotErrorResponse(err error) (int, string) {
	var validation runtimeSnapshotValidationError
	switch {
	case errors.Is(err, ErrAuthorizationForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.As(err, &validation):
		return http.StatusBadRequest, "invalid runtime snapshot"
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "runtime snapshot synchronization timed out"
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout, "runtime snapshot request cancelled"
	default:
		return http.StatusServiceUnavailable, "runtime snapshot synchronization unavailable"
	}
}
