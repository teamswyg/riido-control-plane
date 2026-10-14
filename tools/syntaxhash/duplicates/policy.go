package duplicates

import "strings"

func prefixMatch(path string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func fileCount(groups []Group) int {
	total := 0
	for _, group := range groups {
		total += group.FileCount
	}
	return total
}

func internalGroupCount(groups []Group) int {
	total := 0
	for _, group := range groups {
		for _, pkg := range group.Packages {
			if strings.HasPrefix(pkg, "internal/") {
				total++
				break
			}
		}
	}
	return total
}

func limitGroups(groups []Group, limit int) []Group {
	if limit <= 0 || len(groups) <= limit {
		return groups
	}
	return append([]Group(nil), groups[:limit]...)
}
