package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ReactToAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID, messageID, emoji string, add bool) (threadMessageReactionChange, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return threadMessageReactionChange{}, err
	}
	change, err := s.DevelopmentAIAgentClientStore.ReactToAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, messageID, emoji, add)
	if err != nil || !change.Changed {
		return change, err
	}
	return change, s.saveSnapshot(ctx)
}
