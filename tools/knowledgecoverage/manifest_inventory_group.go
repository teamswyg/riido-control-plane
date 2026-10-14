package main

import (
	"sort"
	"strings"
)

func manifestInventoryByGroup(root string) []manifestDir {
	counts := map[string]int{}
	for _, path := range manifestInventory(root) {
		counts[manifestGroup(path)]++
	}
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

func manifestGroup(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		return "."
	}
	return parts[0]
}
