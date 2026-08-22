// Package controlplanehealth owns the application use case behind the
// control-plane GraphQL healthCheck field.
package controlplanehealth

import (
	"context"
	"net/http"
)

const HealthyStatus = http.StatusOK

type Checker interface {
	HealthCheck(context.Context) (int, error)
}

type Service struct{}

func NewService() Service { return Service{} }

func (Service) HealthCheck(context.Context) (int, error) { return HealthyStatus, nil }
