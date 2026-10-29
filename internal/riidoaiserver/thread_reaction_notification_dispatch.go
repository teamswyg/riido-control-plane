package riidoaiserver

import (
	"context"
	"log"
)

// queueThreadReactionNotification sends member reaction notifications one at
// a time per message and actor, and always sends the latest one by store
// revision. A handler that enqueues late can therefore never send an older
// add after a later removal. Ordering holds within one instance only.
func (s Server) queueThreadReactionNotification(notification agentThreadReactionNotification) {
	d := s.commentReactions
	if d == nil || !s.config.CommentDelivery.monolithEnabled() {
		return
	}
	key := "notify:" + notification.ThreadID + "/" + notification.MessageID + "/" + notification.ActorPrincipalID
	d.mu.Lock()
	queued, waiting := d.notifications[key]
	if notification.revision < d.notificationSent[key] || (waiting && notification.revision < queued.revision) {
		d.mu.Unlock()
		return
	}
	d.notifications[key] = notification
	d.mu.Unlock()
	d.start(key, func() {
		d.mu.Lock()
		latest, ok := d.notifications[key]
		delete(d.notifications, key)
		if ok {
			d.notificationSent[key] = latest.revision
		}
		d.mu.Unlock()
		if !ok {
			return
		}
		if err := s.config.CommentDelivery.postThreadReactionNotification(context.Background(), latest); err != nil {
			log.Printf("agent thread reaction notification failed message=%s: %v", latest.MessageID, err)
		}
	})
}
