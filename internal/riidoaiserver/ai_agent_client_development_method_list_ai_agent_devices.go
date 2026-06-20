package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error) {
	if err := ctx.Err(); err != nil {
		return DeviceRuntimeListResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return DeviceRuntimeListResponse{SchemaVersion: SchemaVersion, Devices: s.visibleDevicesLocked(principal)}, nil
}
