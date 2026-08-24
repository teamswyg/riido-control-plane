// Package controlplanehealth owns the application use case behind the
// control-plane GraphQL healthCheck field.
package controlplanehealth

import (
	"context"
	"errors"
	"net/http"
)

const HealthyStatus = http.StatusOK

var ErrFireErrorAlwaysFails = errors.New("control-plane fireError always fails by contract")

type Checker interface {
	HealthCheck(context.Context) (int, error)
}

type FireErrorer interface {
	FireError(context.Context) error
}

type Service struct{}

func NewService() Service { return Service{} }

func (Service) HealthCheck(context.Context) (int, error) { return HealthyStatus, nil }

// AlwaysFailFireError is the explicit owner composition for the pinned
// NEVER_RETURNING_VOID source operation. It has no successful value path.
type AlwaysFailFireError struct{}

func NewAlwaysFailFireError() AlwaysFailFireError { return AlwaysFailFireError{} }

func (AlwaysFailFireError) FireError(context.Context) error { return ErrFireErrorAlwaysFails }
