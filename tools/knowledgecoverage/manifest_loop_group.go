package main

import "sort"

func manifestLoopGroups(counts map[string]int) []manifestDir {
	groups := make([]manifestDir, 0, len(counts))
	for group, count := range counts {
		groups = append(groups, manifestDir{Group: group, Count: count})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Count == groups[j].Count {
			return groups[i].Group < groups[j].Group
		}
		return groups[i].Count > groups[j].Count
	})
	return groups
}
