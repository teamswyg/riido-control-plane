package main

import "slices"

func topLineBudgetSamples(samples []lineBudgetSample, limit int) []lineBudgetSample {
	slices.SortFunc(samples, func(a, b lineBudgetSample) int { return b.Lines - a.Lines })
	if limit > 0 && len(samples) > limit {
		return samples[:limit]
	}
	return samples
}
