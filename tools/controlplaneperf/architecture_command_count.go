package main

func architectureTargetCommandCount(rows []architectureFileEvidence) int {
	seen := map[string]bool{}
	for _, row := range rows {
		for _, command := range row.TargetVerifierCommands {
			seen[command] = true
		}
	}
	return len(seen)
}
