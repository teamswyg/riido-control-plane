package riidoaiserver

func progressLinesAreStrictlyAfter(existing, incoming []AgentThreadProgressLine) bool {
	previous := maxProgressLineSeq(existing)
	for _, line := range incoming {
		if progressLineReplacesPrevious(line) || line.Seq <= previous {
			return false
		}
		previous = line.Seq
	}
	return true
}

func maxProgressLineSeq(lines []AgentThreadProgressLine) int {
	maxSeq := 0
	for _, line := range lines {
		if line.Seq > maxSeq {
			maxSeq = line.Seq
		}
	}
	return maxSeq
}
