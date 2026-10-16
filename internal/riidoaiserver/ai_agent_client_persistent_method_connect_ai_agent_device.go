package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ConnectAIAgentDevice(ctx context.Context, principal AuthorizationResult, machineID string) (DeviceRecord, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceRecord{}, err
	}
	grantsBefore := s.desktopPresenceGrantCheckpoint()
	device, err := s.DevelopmentAIAgentClientStore.ConnectAIAgentDevice(ctx, principal, machineID)
	if err != nil {
		return device, err
	}
	return device, s.commitDesktopPresenceGrants(ctx, grantsBefore)
}
