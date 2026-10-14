package main

func selectedRefreshLoops(commands []selectedRefreshCommand, expiresAt string) []selectedRefreshLoop {
	counts := map[string]int{}
	order := []string{}
	for _, command := range commands {
		if _, ok := counts[command.LoopID]; !ok {
			order = append(order, command.LoopID)
		}
		counts[command.LoopID]++
	}
	loops := make([]selectedRefreshLoop, 0, len(order))
	for _, loopID := range order {
		loops = append(loops, selectedRefreshLoop{
			LoopID:            loopID,
			EvidenceExpiresAt: expiresAt,
			CommandCount:      counts[loopID],
		})
	}
	return loops
}
