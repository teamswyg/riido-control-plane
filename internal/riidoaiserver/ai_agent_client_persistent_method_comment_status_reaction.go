package riidoaiserver

import "context"

// CommentStatusReactionPlans is inherited without a reload: it reads the state
// this instance just recorded. Recording a write reloads fresh state, so the
// plan read right after it sees a state change committed by another instance.

func (s *PersistentAIAgentClientStore) RecordCommentDeliveryEvent(ctx context.Context, workspaceID, threadID, eventID string) error {
	return s.saveThreadReactionUpdate(ctx, threadID, func() (bool, error) {
		return s.recordCommentDeliveryEvent(ctx, workspaceID, threadID, eventID)
	})
}

func (s *PersistentAIAgentClientStore) MarkCommentStatusReactionApplied(ctx context.Context, workspaceID, threadID, emoji string) error {
	return s.saveThreadReactionUpdate(ctx, threadID, func() (bool, error) {
		return s.markCommentStatusReactionApplied(ctx, workspaceID, threadID, emoji)
	})
}
