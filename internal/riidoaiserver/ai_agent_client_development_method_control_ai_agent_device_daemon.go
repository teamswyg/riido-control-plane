package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ControlAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return DeviceDaemonCommandResponse{}, errors.New("device_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	daemon, ok := s.deviceDaemonForOwnerLocked(principal, deviceID)
	if !ok {
		return DeviceDaemonCommandResponse{}, ErrAIAgentNotFound
	}
	if !daemonSupportsAction(daemon, action) {
		return DeviceDaemonCommandResponse{}, errors.New("daemon action is not supported in the current state")
	}
	result, err := s.applyDaemonControlLocked(daemon, action)
	if err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	return DeviceDaemonCommandResponse{
		SchemaVersion: SchemaVersion,
		CommandID:     result.CommandID,
		DeviceID:      result.Daemon.DeviceID,
		Action:        action,
		Availability:  result.Daemon.Availability,
		ControlState:  result.Daemon.ControlState,
		AcceptedAt:    result.AcceptedAt,
		Message:       result.Message + " for device " + result.Daemon.DeviceID,
	}, nil
}
