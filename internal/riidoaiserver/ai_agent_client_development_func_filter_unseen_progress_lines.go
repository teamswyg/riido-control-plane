package riidoaiserver

func filterUnseenProgressLines(existing, incoming []AgentThreadProgressLine) []AgentThreadProgressLine {
	if len(existing) == 0 || len(incoming) == 0 {
		return incoming
	}
	if progressLinesAreStrictlyAfter(existing, incoming) {
		return incoming
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
