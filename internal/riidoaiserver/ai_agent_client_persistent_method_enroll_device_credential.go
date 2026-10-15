package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) EnrollDeviceCredential(ctx context.Context, principal AuthorizationResult, workspaceID string, req EnrollDeviceRequest) (EnrollDeviceResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return EnrollDeviceResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.EnrollDeviceCredential(ctx, principal, workspaceID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
