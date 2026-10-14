package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) RecordAIAgentAssignmentEvent(ctx context.Context, agentID string, req AgentEventRequest, event TaskEvent) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	if err := s.DevelopmentAIAgentClientStore.RecordAIAgentAssignmentEvent(ctx, agentID, req, event); err != nil {
		return err
	}
	return s.saveSnapshot(ctx)
}
