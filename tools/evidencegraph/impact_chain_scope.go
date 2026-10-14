package main

import "sort"

func refPaths(refs []ref) []string {
	seen := map[string]bool{}
	for _, item := range refs {
		if item.Path == "" {
			continue
		}
		seen[item.Path] = true
	}
	out := make([]string, 0, len(seen))
	for path := range seen {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func sortedValues(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
