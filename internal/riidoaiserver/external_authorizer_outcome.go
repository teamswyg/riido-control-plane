package riidoaiserver

import (
	"context"
	"errors"
)

func externalAuthorizerRetryable(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) ||
		(err != nil && !errors.Is(err, ErrAuthorizationForbidden) &&
			!errors.Is(err, ErrAuthorizationUnauthenticated) && !errors.Is(err, context.Canceled))
}

func externalAuthorizerOutcome(err error) string {
	switch {
	case err == nil:
		return "allowed"
	case errors.Is(err, ErrAuthorizationForbidden):
		return "forbidden"
	case errors.Is(err, ErrAuthorizationUnauthenticated):
		return "unauthenticated"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	default:
		return "error"
	}
}

func externalAuthorizerLogFailure(err error) bool {
	return err != nil && !errors.Is(err, ErrAuthorizationForbidden) &&
		!errors.Is(err, ErrAuthorizationUnauthenticated) && !errors.Is(err, context.Canceled)
}
