package riidoaiserver

func copyClientVisibleProgressLines(lines []AgentThreadProgressLine) []AgentThreadProgressLine {
	out := copyProgressLines(lines)
	for i := range out {
		out[i].Message = clientVisibleTaskThreadText(out[i].Message)
	}
	return out
}
