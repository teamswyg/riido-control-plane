package riidoaiserver

import (
	"context"
	"errors"
	"io"
	"math/rand/v2"
	"net"
	"syscall"
	"time"
)

// Retain the sanitized failure classification without exposing upstream details.
// Only the read-scope HTTP path opts into retrying this marker.
type transientTaskContextReadError struct{ cause error }

func (e transientTaskContextReadError) Error() string { return e.cause.Error() }
func (e transientTaskContextReadError) Unwrap() error { return e.cause }

func isTransientTaskContextRead(err error) bool {
	var transient transientTaskContextReadError
	return errors.As(err, &transient)
}

func transientTaskContextTransport(err error) bool {
	var timeout net.Error
	return errors.Is(err, context.DeadlineExceeded) ||
		(errors.As(err, &timeout) && timeout.Timeout()) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED)
}

// Retry only an explicitly transient read, once, without resetting its budget.
func retryTaskContextRead(ctx context.Context, read func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := read()
	if !isTransientTaskContextRead(err) {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	timer := time.NewTimer(50*time.Millisecond + time.Duration(rand.Int64N(int64(50*time.Millisecond))))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return read()
}
