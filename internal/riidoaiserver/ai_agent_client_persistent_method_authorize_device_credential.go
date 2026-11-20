package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) AuthorizeDeviceCredential(ctx context.Context, deviceID, deviceSecret string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := s.lockOperationContext(ctx); err != nil {
		return AuthorizationResult{}, err
	}
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AuthorizationResult{}, err
	}
	return s.DevelopmentAIAgentClientStore.AuthorizeDeviceCredential(ctx, deviceID, deviceSecret, req)
}
