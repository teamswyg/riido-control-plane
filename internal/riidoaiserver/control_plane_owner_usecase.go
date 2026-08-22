package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
)

var errControlPlaneOwnerFire = errors.New("control-plane fireError always fails")

type controlPlaneOwnerUseCase interface {
	HealthCheck(context.Context) (int, error)
	FireError(context.Context) error
}

type sourceReadyControlPlaneOwner struct{}

func (sourceReadyControlPlaneOwner) HealthCheck(context.Context) (int, error) {
	return http.StatusOK, nil
}

func (sourceReadyControlPlaneOwner) FireError(context.Context) error {
	return errControlPlaneOwnerFire
}
