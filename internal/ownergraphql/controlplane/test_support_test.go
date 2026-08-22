package controlplane

import (
	"context"
	"errors"
)

type controlPlaneOwnerSpy struct {
	healthStatus int
	healthErr    error
	fireErr      error
	healthCalls  int
	fireCalls    int
}

func (s *controlPlaneOwnerSpy) HealthCheck(context.Context) (int, error) {
	s.healthCalls++
	return s.healthStatus, s.healthErr
}

func (s *controlPlaneOwnerSpy) FireError(context.Context) error {
	s.fireCalls++
	return s.fireErr
}

func healthyControlPlaneOwnerSpy() *controlPlaneOwnerSpy {
	return &controlPlaneOwnerSpy{healthStatus: 200, fireErr: errors.New("source always throws")}
}
