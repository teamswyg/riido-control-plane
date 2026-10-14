package main

import "sort"

func countDecisionValues(
	decisions []decisionRecord,
	valueFor func(decisionRecord) string,
) []summaryCount {
	counts := map[string]int{}
	for _, decision := range decisions {
		counts[valueFor(decision)]++
	}
	return sortedSummaryCounts(counts)
}

func sortedSummaryCounts(counts map[string]int) []summaryCount {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]summaryCount, 0, len(keys))
	for _, key := range keys {
		out = append(out, summaryCount{Key: key, Count: counts[key]})
	}
	return out
}
