package riidoaiserver

func mergeThreadProgressLines(existing, incoming []AgentThreadProgressLine) []AgentThreadProgressLine {
	out := existing
	for _, line := range incoming {
		if progressLineReplacesPrevious(line) {
			replaced := false
			for i := len(out) - 1; i >= 0; i-- {
				if progressLineReplacesPrevious(out[i]) {
					out[i] = line
					replaced = true
					break
				}
			}
			if replaced {
				continue
			}
		}
		out = append(out, line)
	}
	return out
}
