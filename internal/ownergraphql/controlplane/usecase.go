package controlplane

import (
	"context"
	"errors"
	"net/http"
)

var errControlPlaneOwnerFire = errors.New("control-plane fireError always fails")

// UseCase is the exact execution boundary for the frozen owner contract.
type UseCase interface {
	HealthCheck(context.Context) (int, error)
	FireError(context.Context) error
}

type runtimeUseCase struct{}

// NewRuntimeUseCase returns the authoritative behavior used by the registered receiver.
func NewRuntimeUseCase() UseCase {
	return runtimeUseCase{}
}

func (runtimeUseCase) HealthCheck(context.Context) (int, error) {
	return http.StatusOK, nil
}

func (runtimeUseCase) FireError(context.Context) error {
	return errControlPlaneOwnerFire
}
