package duplicates

import "sort"

func sortedGroups(index map[string]*Group) []Group {
	groups := []Group{}
	for _, group := range index {
		group.Packages, group.Files = uniqueSorted(group.Packages), uniqueSorted(group.Files)
		group.FileCount = len(group.Files)
		if group.FileCount > 1 {
			groups = append(groups, *group)
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].FileCount != groups[j].FileCount {
			return groups[i].FileCount > groups[j].FileCount
		}
		return groups[i].ShapeHash < groups[j].ShapeHash
	})
	return groups
}

func uniqueSorted(values []string) []string {
	sort.Strings(values)
	out := values[:0]
	for _, value := range values {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}
