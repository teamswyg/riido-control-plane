package main

import "sort"

func architecturePathBindings(
	byPath map[string]*architecturePathBinding,
) []architecturePathBinding {
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	out := make([]architecturePathBinding, 0, len(paths))
	for _, path := range paths {
		out = append(out, *byPath[path])
	}
	return out
}
