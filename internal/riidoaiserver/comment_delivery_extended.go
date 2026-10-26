package riidoaiserver

import "strings"

// commentDeliveryExtendedContext keeps the delivered comment as the
// priority-1 comment. Delivery carries no user credential, so related items
// and other comments cannot be read for this run.
func commentDeliveryExtendedContext(commentID, body string) *AIAgentTaskContextExtended {
	extended := &AIAgentTaskContextExtended{RelatedStatus: "unavailable", CommentStatus: "partial"}
	if commentID = strings.TrimSpace(commentID); commentID != "" {
		extended.Comments = []AIAgentTaskContextComment{{ID: commentID, Priority: TaskContextCommentPriorityTrigger, Body: boundedTaskContextText(body, taskContextCommentBodyMax)}}
		extended.CommentTotal = 1
	}
	return extended
}
