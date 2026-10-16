package riidoaiserver

import "context"

// Callers hold operationMu across checkpoint, mutation, commit and rollback.
// Roll back only access facts; unrelated daemon heartbeat state is untouched.
func (s *PersistentAIAgentClientStore) desktopPresenceGrantCheckpoint() []DeviceConnectionGrant {
	s.mu.Lock()
	defer s.mu.Unlock()
	return copyDeviceConnectionGrants(s.deviceConnectionGrants)
}

func (s *PersistentAIAgentClientStore) commitDesktopPresenceGrants(ctx context.Context, before []DeviceConnectionGrant) error {
	if err := s.saveSnapshot(ctx); err != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		restored := map[string]map[string]DeviceConnectionGrant{}
		for _, grant := range before {
			if restored[grant.DeviceID] == nil {
				restored[grant.DeviceID] = map[string]DeviceConnectionGrant{}
			}
			restored[grant.DeviceID][deviceConnectionGrantKey(grant.DeviceID, grant.PrincipalID, grant.WorkspaceID)] = grant
		}
		s.deviceConnectionGrants = restored
		return err
	}
	return nil
}
