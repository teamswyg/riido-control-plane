package riidoaiserver

import (
	"context"
	"errors"
)

func taskContextHTTPFailure(status int) error {
	outcome := "unavailable"
	switch {
	case status == 401 || status == 403:
		outcome = "denied"
	case status == 404:
		outcome = "missing"
	case status >= 400 && status < 500:
		outcome = "invalid"
	}
	return taskContextReadError{outcome}
}

func taskContextTransportFailure(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	return taskContextReadError{"unavailable"}
}
