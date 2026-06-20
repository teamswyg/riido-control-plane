package riidoaiserver

func normalizeProgressLines(lines []AgentThreadProgressLine) []AgentThreadProgressLine {
	out := make([]AgentThreadProgressLine, 0, len(lines))
	for _, line := range lines {
		normalized, ok := normalizeProgressLine(line)
		if !ok {
			continue
		}
		if normalized.Seq <= 0 {
			normalized.Seq = len(out) + 1
		}
		out = append(out, normalized)
	}
	return out
}
