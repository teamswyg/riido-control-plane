package riidoaiserver

func filterUnseenProgressLines(existing, incoming []AgentThreadProgressLine) []AgentThreadProgressLine {
	if len(existing) == 0 || len(incoming) == 0 {
		return incoming
	}
	if progressLinesAreStrictlyAfter(existing, incoming) {
		return incoming
	}
	if len(incoming) <= 8 {
		return filterSmallUnseenProgressLines(existing, incoming)
	}
	seen := make(map[int]struct{}, len(existing))
	for _, line := range existing {
		if line.Seq > 0 {
			seen[line.Seq] = struct{}{}
		}
	}
	out := incoming[:0]
	for _, line := range incoming {
		if progressLineReplacesPrevious(line) {
			out = append(out, line)
			continue
		}
		if line.Seq > 0 {
			if _, ok := seen[line.Seq]; ok {
				continue
			}
			seen[line.Seq] = struct{}{}
		}
		out = append(out, line)
	}
	return out
}

func filterSmallUnseenProgressLines(existing, incoming []AgentThreadProgressLine) []AgentThreadProgressLine {
	out := incoming[:0]
	for i, line := range incoming {
		if progressLineReplacesPrevious(line) {
			out = append(out, line)
			continue
		}
		if line.Seq > 0 && (progressLineSeqSeen(existing, line.Seq) ||
			progressIncomingSeqSeen(incoming[:i], line.Seq)) {
			continue
		}
		out = append(out, line)
	}
	return out
}

func progressIncomingSeqSeen(lines []AgentThreadProgressLine, seq int) bool {
	for _, line := range lines {
		if line.Seq == seq && !progressLineReplacesPrevious(line) {
			return true
		}
	}
	return false
}
