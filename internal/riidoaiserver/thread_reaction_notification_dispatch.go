package riidoaiserver

import (
	"context"
	"log"
)

// queueThreadReactionNotification sends member reaction notifications one at
// a time per message and actor, and always sends the latest one. A delayed add
// can therefore never recreate a notification after a later removal.
func (s Server) queueThreadReactionNotification(notification agentThreadReactionNotification) {
	d := s.commentReactions
	if d == nil || !s.config.CommentDelivery.monolithEnabled() {
		return
	}
	key := "notify:" + notification.ThreadID + "/" + notification.MessageID + "/" + notification.ActorPrincipalID
	d.mu.Lock()
	d.notifications[key] = notification
	d.mu.Unlock()
	d.start(key, func() {
		d.mu.Lock()
		latest, ok := d.notifications[key]
		delete(d.notifications, key)
		d.mu.Unlock()
		if !ok {
			return
		}
		if err := s.config.CommentDelivery.postThreadReactionNotification(context.Background(), latest); err != nil {
			log.Printf("agent thread reaction notification failed message=%s: %v", latest.MessageID, err)
		}
	})
}
