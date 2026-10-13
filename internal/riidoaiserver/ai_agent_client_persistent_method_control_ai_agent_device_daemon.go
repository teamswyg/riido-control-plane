package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ControlAIAgentDeviceDaemon(ctx context.Context, principal AuthorizationResult, deviceID string, action DaemonControlAction, req ControlDeviceDaemonRequest) (DeviceDaemonCommandResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceDaemonCommandResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.ControlAIAgentDeviceDaemon(ctx, principal, deviceID, action, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
