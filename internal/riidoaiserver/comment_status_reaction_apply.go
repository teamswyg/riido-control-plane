package riidoaiserver

import (
	"context"
	"errors"
	"log"
)

// commentStatusReactionPasses bounds the writes of one dispatch. After each
// write the plan is read again, so a state change that arrived while the PUT
// was in flight (a stop during the first 👀) is written in the same dispatch.
const commentStatusReactionPasses = 4

func (s Server) applyCommentStatusReaction(store commentStatusReactionStore, taskID, threadID string) {
	for range commentStatusReactionPasses {
		plan, ok := commentStatusReactionPlanFor(store.CommentStatusReactionPlans(taskID), threadID)
		if !ok || !s.writeCommentStatusReaction(store, plan) {
			return
		}
	}
}

func commentStatusReactionPlanFor(plans []commentStatusReactionPlan, threadID string) (commentStatusReactionPlan, bool) {
	for _, plan := range plans {
		if plan.ThreadID == threadID {
			return plan, true
		}
	}
	return commentStatusReactionPlan{}, false
}

// writeCommentStatusReaction reports whether the write was recorded. A
// permanent rejection is recorded too, so a deleted comment is not retried.
func (s Server) writeCommentStatusReaction(store commentStatusReactionStore, plan commentStatusReactionPlan) bool {
	ctx := context.Background()
	err := s.config.CommentDelivery.putCommentStatusReaction(ctx, plan.EventID, plan.Emoji)
	var statusErr monolithStatusError
	if err != nil && (!errors.As(err, &statusErr) || !statusErr.permanent()) {
		log.Printf("agent comment status reaction failed thread=%s: %v", plan.ThreadID, err)
		return false
	}
	if err := store.MarkCommentStatusReactionApplied(ctx, plan.WorkspaceID, plan.ThreadID, plan.Emoji); err != nil {
		log.Printf("agent comment status reaction not recorded thread=%s: %v", plan.ThreadID, err)
		return false
	}
	return true
}
