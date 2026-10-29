package riidoaiserver

import (
	"context"
	"log"
)

// saveAssignmentContextComments stores comment pages before the snapshot so a
// snapshot never points at missing pages; on failure comments are reported
// unavailable instead.
func (s Server) saveAssignmentContextComments(ctx context.Context, snapshot *AssignmentContextSnapshot, comments []AIAgentTaskContextComment) int {
	pages := pageAssignmentContextComments(comments)
	if len(pages) == 0 {
		return 0
	}
	if err := s.config.AssignmentContext.SaveAssignmentContextComments(ctx, snapshot.AssignmentID, pages); err != nil {
		log.Printf("event=assignment_context_capture_failed assignment_id=%q stage=comments error_type=%T", snapshot.AssignmentID, err)
		if snapshot.Context.Extended != nil {
			extended := *snapshot.Context.Extended
			extended.CommentStatus = "unavailable"
			snapshot.Context.Extended = &extended
		}
		return 0
	}
	return len(pages)
}
