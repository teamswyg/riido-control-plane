package riidoaiserver

import (
	"strconv"
	"strings"
)

func aiAgentClientReplayEventsAfterLastEventID(
	events []ClientStreamEvent,
	lastEventID string,
) []ClientStreamEvent {
	cursor, err := strconv.ParseInt(strings.TrimSpace(lastEventID), 10, 64)
	if err != nil || cursor < 0 {
		return events
	}
	for i, event := range events {
		if event.Seq > cursor {
			return events[i:]
		}
	}
	return events[:0]
}
