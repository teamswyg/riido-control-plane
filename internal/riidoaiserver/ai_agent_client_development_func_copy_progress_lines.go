package riidoaiserver

func copyProgressLines(lines []AgentThreadProgressLine) []AgentThreadProgressLine {
	return append([]AgentThreadProgressLine(nil), lines...)
}
