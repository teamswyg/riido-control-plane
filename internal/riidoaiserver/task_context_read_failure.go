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
	failure := taskContextReadError{outcome}
	if status == 502 || status == 503 || status == 504 {
		return transientTaskContextReadError{failure}
	}
	return failure
}

func taskContextTransportFailure(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return transientTaskContextReadError{context.DeadlineExceeded}
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	failure := taskContextReadError{"unavailable"}
	if transientTaskContextTransport(err) {
		return transientTaskContextReadError{failure}
	}
	return failure
}
