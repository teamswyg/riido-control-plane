package riidoaiserver

func retainLatestThreadProgressLines(lines []AgentThreadProgressLine) []AgentThreadProgressLine {
	if len(lines) <= aiAgentClientThreadProgressLineLimit {
		return lines
	}
	retained := make([]AgentThreadProgressLine, aiAgentClientThreadProgressLineLimit)
	copy(retained, lines[len(lines)-aiAgentClientThreadProgressLineLimit:])
	return retained
}
