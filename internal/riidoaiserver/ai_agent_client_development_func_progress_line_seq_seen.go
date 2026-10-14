package riidoaiserver

func progressLineSeqSeen(lines []AgentThreadProgressLine, seq int) bool {
	if seq <= 0 {
		return false
	}
	for _, line := range lines {
		if line.Seq == seq {
			return true
		}
	}
	return false
}
