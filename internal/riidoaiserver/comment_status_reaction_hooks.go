package riidoaiserver

import (
	"context"
	"log"
	"net/http"
	"strings"
)

// acceptCommentDelivery answers an accepted delivery after linking it to its
// run thread for the comment status reaction.
func (s Server) acceptCommentDelivery(ctx context.Context, w http.ResponseWriter, evidence verifiedCommentDelivery, response AIAgentTaskActionResponse) {
	s.trackCommentDeliveryReaction(ctx, evidence.WorkspaceID, evidence.TaskID, response.ThreadID, evidence.EventID)
	writeJSON(w, http.StatusAccepted, response)
}

// trackCommentDeliveryReaction links an accepted delivery to its run thread and
// writes the started reaction. A failure never fails delivery; the link is
// retried in the background without running the agent again.
func (s Server) trackCommentDeliveryReaction(ctx context.Context, workspaceID, taskID, threadID, eventID string) {
	store, ok := s.aiAgent.(commentStatusReactionStore)
	if !ok || strings.TrimSpace(threadID) == "" {
		return
	}
	if err := store.RecordCommentDeliveryEvent(ctx, workspaceID, threadID, eventID); err != nil {
		log.Printf("agent comment delivery reaction source not recorded thread=%s: %v", threadID, err)
		s.retryCommentDeliveryLink(store, workspaceID, taskID, threadID, eventID)
		return
	}
	s.reconcileCommentStatusReactions(taskID)
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
