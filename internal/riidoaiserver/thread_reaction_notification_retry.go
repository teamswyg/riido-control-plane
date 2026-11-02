package riidoaiserver

import (
	"time"
)

const memberReactionNotificationMaxAttempts = 4

// memberReactionNotificationRetryDelay is the wait before the first retry; it
// doubles for each later retry. Tests shorten it.
var memberReactionNotificationRetryDelay = time.Second

type memberReactionNotification struct {
	key      string
	revision uint64
	attempt  int
	body     map[string]string
}
