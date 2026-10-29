package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) RecordAIAgentContextReads(ctx context.Context, reads []AssignmentContextRead) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	if err := s.DevelopmentAIAgentClientStore.RecordAIAgentContextReads(ctx, reads); err != nil {
		return err
	}
	return s.saveSnapshot(ctx)
}
