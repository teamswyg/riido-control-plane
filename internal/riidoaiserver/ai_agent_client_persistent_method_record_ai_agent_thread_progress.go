package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) RecordAIAgentThreadProgress(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) (AgentThreadProgressBatchResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentThreadProgressBatchResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.RecordAIAgentThreadProgress(ctx, agentID, req)
	if err != nil {
		return response, err
	}
	if response.AcceptedLines == 0 {
		return response, nil
	}
	return response, s.saveSnapshot(ctx)
}
