package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ListAIAgentDeviceDaemons(ctx context.Context, principal AuthorizationResult, deviceID string) (DeviceDaemonListResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceDaemonListResponse{}, err
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return DeviceDaemonListResponse{}, errors.New("device_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	daemons, ok := s.visibleDaemonsForDeviceLocked(principal, deviceID)
	if !ok {
		return DeviceDaemonListResponse{}, ErrAIAgentNotFound
	}
	return DeviceDaemonListResponse{SchemaVersion: SchemaVersion, DeviceID: deviceID, Daemons: daemons}, nil
}
