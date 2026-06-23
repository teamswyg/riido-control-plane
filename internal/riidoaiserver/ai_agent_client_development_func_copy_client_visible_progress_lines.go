package riidoaiserver

func copyClientVisibleProgressLines(lines []AgentThreadProgressLine) []AgentThreadProgressLine {
	out := make([]AgentThreadProgressLine, 0, len(lines))
	for _, line := range copyProgressLines(lines) {
		line.Message = clientVisibleTaskThreadText(line.Message)
		if line.Message == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}
