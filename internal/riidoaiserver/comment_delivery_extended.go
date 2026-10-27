package riidoaiserver

import "strings"

// commentDeliveryExtendedContext is the fallback for legacy evidence without
// extended context. Delivery carries no user credential for additional reads;
// newer evidence is enriched by monolith under the verified author's scope.
func commentDeliveryExtendedContext(commentID, body string) *AIAgentTaskContextExtended {
	extended := &AIAgentTaskContextExtended{RelatedStatus: "unavailable", CommentStatus: "partial"}
	if commentID = strings.TrimSpace(commentID); commentID != "" {
		extended.Comments = []AIAgentTaskContextComment{{ID: commentID, Priority: TaskContextCommentPriorityTrigger, Body: boundedTaskContextText(body, taskContextCommentBodyMax)}}
		extended.CommentTotal = 1
	}
	return extended
}
