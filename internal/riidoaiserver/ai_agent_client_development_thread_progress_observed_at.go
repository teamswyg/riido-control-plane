package riidoaiserver

import "time"

func stampMissingProgressObservedAt(lines []AgentThreadProgressLine, observedAt time.Time) []AgentThreadProgressLine {
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	for i := range lines {
		if lines[i].ObservedAt.IsZero() {
			lines[i].ObservedAt = observedAt
		}
	}
	return lines
}
