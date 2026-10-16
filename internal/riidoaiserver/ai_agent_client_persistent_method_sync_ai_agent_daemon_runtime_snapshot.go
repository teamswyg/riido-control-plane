package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) SyncAIAgentDaemonRuntimeSnapshot(ctx context.Context, principal AuthorizationResult, req DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, error) {
	if err := s.lockOperationContext(ctx); err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, err
	}
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceRuntimeSnapshotSyncResponse{}, err
	}
	response, changed, err := s.syncAIAgentDaemonRuntimeSnapshot(ctx, principal, req)
	if err != nil {
		if changed {
			if saveErr := s.saveSnapshot(ctx); saveErr != nil {
				return response, saveErr
			}
		}
		return response, err
	}
	if !changed && !s.shouldSaveHeartbeatSnapshot() {
		return response, nil
	}
	return response, s.saveSnapshot(ctx)
}
