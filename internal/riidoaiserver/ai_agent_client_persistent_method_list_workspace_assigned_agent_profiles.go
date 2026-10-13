package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListWorkspaceAssignedAgentProfiles(ctx context.Context, principal AuthorizationResult) (AssignedAgentProfileMapResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AssignedAgentProfileMapResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListWorkspaceAssignedAgentProfiles(ctx, principal)
}
