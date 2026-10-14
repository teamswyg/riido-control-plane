package main

func sumNextLoopChains(items []graphNextLoopSummary) int {
	total := 0
	for _, item := range items {
		total += item.ChainCount
	}
	return total
}
