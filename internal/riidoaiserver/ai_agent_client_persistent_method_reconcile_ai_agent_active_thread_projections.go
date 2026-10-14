package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ReconcileAIAgentActiveThreadProjections(ctx context.Context, principal AuthorizationResult, taskID string, reader AssignmentProjectionReader) (bool, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return false, err
	}
	changed, err := s.DevelopmentAIAgentClientStore.ReconcileAIAgentActiveThreadProjections(ctx, principal, taskID, reader)
	if err != nil || !changed {
		return changed, err
	}
	return changed, s.saveSnapshot(ctx)
}
