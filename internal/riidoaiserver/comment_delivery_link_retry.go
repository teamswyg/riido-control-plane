package riidoaiserver

import (
	"context"
	"log"
	"time"
)

// retryCommentDeliveryLink records the delivery event of a comment-started run
// again in the background after a failed first try, with the same doubling
// delay and attempt bound as member reaction notifications. Without the link
// the comment would never receive a status reaction.
func (s Server) retryCommentDeliveryLink(store commentStatusReactionStore, workspaceID, taskID, threadID, eventID string) {
	if s.commentReactions == nil {
		return
	}
	s.commentReactions.start("link:"+threadID, func() {
		for attempt := 1; attempt < threadReactionNotificationMaxAttempts; attempt++ {
			time.Sleep(threadReactionNotificationRetryDelay << (attempt - 1))
			err := store.RecordCommentDeliveryEvent(context.Background(), workspaceID, threadID, eventID)
			if err == nil {
				s.reconcileCommentStatusReactions(taskID)
				return
			}
			log.Printf("agent comment delivery reaction source retry failed thread=%s attempt=%d: %v", threadID, attempt+1, err)
		}
	})
}
