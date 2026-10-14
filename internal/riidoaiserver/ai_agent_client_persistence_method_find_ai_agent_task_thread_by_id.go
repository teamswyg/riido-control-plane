package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) FindAIAgentTaskThreadByID(ctx context.Context, workspaceID, threadID string) (AIAgentTaskThreadRecord, error) {
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadRecord{}, err
	}
	return s.DevelopmentAIAgentClientStore.FindAIAgentTaskThreadByID(ctx, workspaceID, threadID)
}
