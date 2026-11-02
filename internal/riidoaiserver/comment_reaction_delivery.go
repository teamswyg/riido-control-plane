package riidoaiserver

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
)

// Read the current projection rather than the arriving event, which can be a retry
// of an older state. Failed writes are repaired by the next history read.
func (s Server) deliverCommentStatusReactions(ctx context.Context, principal AuthorizationResult, taskID string) {
	if s.config.CommentDelivery.EvidenceBaseURL == "" {
		return
	}
	response, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return
	}
	for _, thread := range response.Threads {
		call := thread.CommentInvocation
		if call == nil || call.EvidenceEventID == "" || call.OriginMessageID != "" {
			continue
		}
		emoji, known := agentStatusReaction(thread)
		if !known || s.commentReactions.alreadySent(call.EvidenceEventID, emoji) {
			continue
		}
		if err := s.writeMonolithReaction(ctx, http.MethodPut, "/internal/agent-comment-deliveries/"+url.PathEscape(call.EvidenceEventID)+"/reaction", map[string]string{"emoji": emoji}); err != nil {
			slog.WarnContext(ctx, "agent comment reaction pending", "assignment_id", thread.AssignmentID, "error", err)
		} else {
			s.commentReactions.markSent(call.EvidenceEventID, emoji)
		}
	}
}
