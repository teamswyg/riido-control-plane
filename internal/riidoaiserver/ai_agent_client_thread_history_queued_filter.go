package riidoaiserver

import "time"

func historyQueuedStatusIsSuperseded(
	message AIAgentTaskThreadHistoryMessage,
	cutoff time.Time,
	hasCutoff bool,
	hasRunning bool,
) bool {
	if !historyMessageIsQueuedStatus(message) {
		return false
	}
	return hasRunning || (hasCutoff && !cutoff.Before(message.ObservedAt))
}
