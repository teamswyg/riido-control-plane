package riidoaiserver

import "context"

// CommentStatusReactionPlans is inherited without a reload: it reads the state
// this instance just recorded, and a stale plan only repeats an idempotent PUT.

func (s *PersistentAIAgentClientStore) RecordCommentDeliveryEvent(ctx context.Context, workspaceID, threadID, eventID string) error {
	return s.saveThreadReactionUpdate(ctx, func() error {
		return s.DevelopmentAIAgentClientStore.RecordCommentDeliveryEvent(ctx, workspaceID, threadID, eventID)
	})
}

func (s *PersistentAIAgentClientStore) MarkCommentStatusReactionApplied(ctx context.Context, workspaceID, threadID, emoji string) error {
	return s.saveThreadReactionUpdate(ctx, func() error {
		return s.DevelopmentAIAgentClientStore.MarkCommentStatusReactionApplied(ctx, workspaceID, threadID, emoji)
	})
}

func (s *PersistentAIAgentClientStore) saveThreadReactionUpdate(ctx context.Context, update func() error) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return err
	}
	if err := update(); err != nil {
		return err
	}
	return s.saveSnapshot(ctx)
}
