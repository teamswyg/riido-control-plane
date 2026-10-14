package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) GetAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string) (DeviceDaemonDetailResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceDaemonDetailResponse{}, err
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return DeviceDaemonDetailResponse{}, errors.New("device_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	daemon, ok := s.deviceDaemonForOwnerLocked(principal, deviceID)
	if !ok {
		return DeviceDaemonDetailResponse{}, ErrAIAgentNotFound
	}
	response := DeviceDaemonDetailResponse{SchemaVersion: SchemaVersion, Daemon: copyDeviceDaemon(daemon)}
	if device, ok := s.deviceByIDLocked(daemon.DeviceID); ok {
		if visibleDevice, ok := s.visibleDeviceRecordLocked(principal, device); ok {
			response.Runtimes = append([]RuntimeRecord(nil), visibleDevice.Runtimes...)
		}
	}
	return response, nil
}
