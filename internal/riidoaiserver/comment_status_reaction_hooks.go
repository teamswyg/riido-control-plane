package riidoaiserver

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
)

// trackCommentDeliveryReaction links an accepted delivery to its run thread and
// writes the started reaction. A failure is logged and never fails delivery.
func (s Server) trackCommentDeliveryReaction(ctx context.Context, workspaceID, taskID, threadID, eventID string) {
	store, ok := s.aiAgent.(commentStatusReactionStore)
	if !ok || strings.TrimSpace(threadID) == "" {
		return
	}
	if err := store.RecordCommentDeliveryEvent(ctx, workspaceID, threadID, eventID); err != nil {
		log.Printf("agent comment delivery reaction source not recorded thread=%s: %v", threadID, err)
		return
	}
	s.reconcileCommentStatusReactions(taskID)
}

// reconcileCommentStatusReactionsAfter runs after task requests that may stop
// or reassign a comment-started run.
func (s Server) reconcileCommentStatusReactionsAfter(r *http.Request, taskID string) {
	if r.Method != http.MethodGet {
		s.reconcileCommentStatusReactions(taskID)
	}
}

// reconcileCommentStatusReactions writes the agent status reaction of each
// comment-started run of the task to monolith. It never blocks or fails the
// caller; a failed write is retried on the next run state change.
func (s Server) reconcileCommentStatusReactions(taskID string) {
	store, ok := s.aiAgent.(commentStatusReactionStore)
	if !ok || s.commentReactions == nil || !s.config.CommentDelivery.monolithEnabled() {
		return
	}
	for _, plan := range store.CommentStatusReactionPlans(taskID) {
		threadID := plan.ThreadID
		s.commentReactions.start(threadID, func() { s.applyCommentStatusReaction(store, taskID, threadID) })
	}
}

func (s Server) applyCommentStatusReaction(store commentStatusReactionStore, taskID, threadID string) {
	for _, plan := range store.CommentStatusReactionPlans(taskID) {
		if plan.ThreadID != threadID {
			continue
		}
		ctx := context.Background()
		err := s.config.CommentDelivery.putCommentStatusReaction(ctx, plan.EventID, plan.Emoji)
		var statusErr monolithStatusError
		if err != nil && (!errors.As(err, &statusErr) || !statusErr.permanent()) {
			log.Printf("agent comment status reaction failed thread=%s: %v", threadID, err)
			return
		}
		if err := store.MarkCommentStatusReactionApplied(ctx, plan.WorkspaceID, threadID, plan.Emoji); err != nil {
			log.Printf("agent comment status reaction not recorded thread=%s: %v", threadID, err)
		}
		return
	}
}
