package controlplane

import (
	"context"
	"errors"
	"net/http"
)

var errControlPlaneOwnerFire = errors.New("control-plane fireError always fails")

// UseCase is the exact source-ready execution boundary for the frozen owner contract.
type UseCase interface {
	HealthCheck(context.Context) (int, error)
	FireError(context.Context) error
}

type sourceReadyUseCase struct{}

// NewSourceReadyUseCase returns the authoritative source behavior without registering a route.
func NewSourceReadyUseCase() UseCase {
	return sourceReadyUseCase{}
}

func (sourceReadyUseCase) HealthCheck(context.Context) (int, error) {
	return http.StatusOK, nil
}

func (sourceReadyUseCase) FireError(context.Context) error {
	return errControlPlaneOwnerFire
}
