package main

import "sort"

func summarizeGraphNextLoops(chains []evidenceChain) []graphNextLoopSummary {
	byLoop := map[string]int{}
	for _, chain := range chains {
		byLoop[chain.NextLoop]++
	}
	out := make([]graphNextLoopSummary, 0, len(byLoop))
	for nextLoop, count := range byLoop {
		out = append(out, graphNextLoopSummary{NextLoop: nextLoop, ChainCount: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ChainCount != out[j].ChainCount {
			return out[i].ChainCount > out[j].ChainCount
		}
		return out[i].NextLoop < out[j].NextLoop
	})
	return out
}
