package riidoaiserver

import (
	"errors"
	"time"
)

const threadReactionNotificationMaxAttempts = 4

// threadReactionNotificationRetryDelay is the wait before the first retry; it
// doubles for each later retry. Tests shorten it.
var threadReactionNotificationRetryDelay = time.Second

// retryThreadReactionNotification queues a failed delivery again unless monolith
// rejected it for good, the attempts ran out, or a newer notification for the
// same key is already waiting. The worker sends it again after the delay.
func (d *commentStatusReactionDispatcher) retryThreadReactionNotification(key string, failed agentThreadReactionNotification, err error) {
	var status monolithStatusError
	if (errors.As(err, &status) && status.permanent()) || failed.attempt+1 >= threadReactionNotificationMaxAttempts {
		return
	}
	d.mu.Lock()
	if _, newer := d.notifications[key]; newer {
		d.mu.Unlock()
		return
	}
	failed.attempt++
	d.notifications[key] = failed
	d.pending[key] = true
	d.mu.Unlock()
	time.Sleep(threadReactionNotificationRetryDelay << (failed.attempt - 1))
}
