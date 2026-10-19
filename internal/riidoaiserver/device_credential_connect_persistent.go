package riidoaiserver

import (
	"context"
	"strings"
)

func (s *PersistentAIAgentClientStore) ConnectAIAgentDeviceWithCredential(ctx context.Context, principal AuthorizationResult, deviceID, secret string) (DeviceRecord, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceRecord{}, err
	}
	before := s.desktopPresenceGrantCheckpoint()
	// A failed commit must not leave a workspace connection for a later save.
	s.mu.Lock()
	var workspacesBefore []string
	for _, device := range s.devices {
		if device.DeviceID == strings.TrimSpace(deviceID) {
			workspacesBefore = append([]string(nil), device.ConnectedWorkspaceIDs...)
			break
		}
	}
	s.mu.Unlock()
	device, err := s.DevelopmentAIAgentClientStore.ConnectAIAgentDeviceWithCredential(ctx, principal, deviceID, secret)
	if err != nil {
		return DeviceRecord{}, err
	}
	if err := s.commitDesktopPresenceGrants(ctx, before); err != nil {
		s.mu.Lock()
		for i := range s.devices {
			if s.devices[i].DeviceID == strings.TrimSpace(deviceID) {
				s.devices[i].ConnectedWorkspaceIDs = workspacesBefore
				break
			}
		}
		s.mu.Unlock()
		return DeviceRecord{}, err
	}
	return device, nil
}
