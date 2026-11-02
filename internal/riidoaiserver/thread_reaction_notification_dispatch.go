package riidoaiserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// queueMemberReactionNotification sends one notification at a time per
// message and member and always sends the latest by store revision, so a
// handler that arrives late never sends an older add after a later removal.
// Ordering holds within one instance only.
func (s Server) queueMemberReactionNotification(notification memberReactionNotification) {
	d := s.commentReactions
	if d == nil || s.config.CommentDelivery.EvidenceBaseURL == "" || s.config.CommentDelivery.Key == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.notifications == nil {
		d.notifications, d.notificationSent, d.notificationRunning = map[string]memberReactionNotification{}, map[string]uint64{}, map[string]bool{}
	}
	key := notification.key
	if queued, waiting := d.notifications[key]; notification.revision < d.notificationSent[key] || (waiting && notification.revision < queued.revision) {
		return
	}
	d.notifications[key] = notification
	if d.notificationRunning[key] {
		return
	}
	d.notificationRunning[key] = true
	d.notificationWG.Go(func() { s.sendMemberReactionNotifications(key) })
}

func (s Server) sendMemberReactionNotifications(key string) {
	d := s.commentReactions
	for {
		d.mu.Lock()
		latest, ok := d.notifications[key]
		if !ok {
			delete(d.notificationRunning, key)
			d.mu.Unlock()
			return
		}
		delete(d.notifications, key)
		if len(d.notificationSent) >= 2048 {
			clear(d.notificationSent)
		}
		d.notificationSent[key] = latest.revision
		d.mu.Unlock()
		err := s.writeMonolithReaction(context.Background(), http.MethodPost, "/internal/agent-thread-reaction-notifications", latest.body)
		if err == nil {
			continue
		}
		slog.Warn("thread reaction notification pending", "message_id", latest.body["message_id"], "attempt", latest.attempt+1, "error", err)
		var status monolithReactionStatusError
		if (errors.As(err, &status) && status.permanent()) || latest.attempt+1 >= memberReactionNotificationMaxAttempts {
			continue
		}
		time.Sleep(memberReactionNotificationRetryDelay << latest.attempt)
		d.mu.Lock()
		if _, newer := d.notifications[key]; !newer {
			latest.attempt++
			d.notifications[key] = latest
		}
		d.mu.Unlock()
	}
}
