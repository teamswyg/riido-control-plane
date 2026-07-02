package riidoaiserver

func appendRetainedThreadProgressLines(existing, incoming []AgentThreadProgressLine) []AgentThreadProgressLine {
	if len(incoming) == 0 {
		return existing
	}
	if !progressLinesAreAppendOnly(incoming) {
		return retainLatestThreadProgressLines(mergeThreadProgressLines(existing, incoming))
	}
	if len(incoming) >= aiAgentClientThreadProgressLineLimit {
		return retainLatestThreadProgressLines(copyProgressLines(incoming))
	}
	if len(existing)+len(incoming) <= aiAgentClientThreadProgressLineLimit {
		return append(existing, incoming...)
	}
	return appendRetainedProgressLinesInPlace(existing, incoming)
}

func progressLinesAreAppendOnly(lines []AgentThreadProgressLine) bool {
	for _, line := range lines {
		if progressLineReplacesPrevious(line) {
			return false
		}
	}
	return true
}

func appendRetainedProgressLinesInPlace(existing, incoming []AgentThreadProgressLine) []AgentThreadProgressLine {
	drop := len(existing) + len(incoming) - aiAgentClientThreadProgressLineLimit
	if drop >= len(existing) {
		start := drop - len(existing)
		copy(existing[:0], incoming[start:])
		return existing[:len(incoming)-start]
	}
	kept := len(existing) - drop
	copy(existing, existing[drop:])
	copy(existing[kept:], incoming)
	return existing[:aiAgentClientThreadProgressLineLimit]
}
