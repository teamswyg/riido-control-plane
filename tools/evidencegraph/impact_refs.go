package main

import "sort"

func changedChainExecutableRefs(item chain, changed map[string]bool) []string {
	paths := map[string]bool{}
	for _, ref := range append(item.Changes, item.Verifiers...) {
		if !isExecutableRef(ref) || !changed[ref.Path] {
			continue
		}
		paths[ref.Path] = true
	}
	out := make([]string, 0, len(paths))
	for path := range paths {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func isExecutableRef(ref ref) bool {
	switch ref.Kind {
	case "code", "test", "tool", "workflow":
		return true
	default:
		return false
	}
}
