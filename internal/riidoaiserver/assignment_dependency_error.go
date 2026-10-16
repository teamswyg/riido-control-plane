package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
)

// Dependency failures must retain retry semantics without exporting upstream
// response bodies as validation errors to the daemon.
type assignmentDependencyError struct{ cause error }

func (e assignmentDependencyError) Error() string { return "assignment dependency unavailable" }
func (e assignmentDependencyError) Unwrap() error { return e.cause }

func assignmentDependencyFailure(err error) error {
	if err == nil {
		return nil
	}
	return assignmentDependencyError{cause: err}
}

func agentEventErrorResponse(err error) (int, string) {
	var dependency assignmentDependencyError
	switch {
	case errors.Is(err, ErrAuthorizationForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, ErrAuthorizationUnauthenticated), errors.Is(err, ErrAuthorizationInvalidCredential):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "assignment event persistence timed out; retry the same event"
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout, "assignment event request cancelled; retry the same event"
	case errors.As(err, &dependency), errors.Is(err, errAssignmentProjectionConflict):
		return http.StatusServiceUnavailable, "assignment event persistence unavailable; retry the same event"
	default:
		return http.StatusBadRequest, err.Error()
	}
}
