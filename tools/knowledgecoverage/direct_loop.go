package main

import "fmt"

func validateDirectLoops(docs []docClass) []string {
	var problems []string
	for _, doc := range docs {
		if doc.Kind == "direct_ssot" && !doc.HasLoop {
			problems = append(problems, fmt.Sprintf("%s direct SSOT manifest must define a complete evidence loop", doc.Path))
		}
	}
	return problems
}

func countDirectLoops(docs []docClass) int {
	count := 0
	for _, doc := range docs {
		if doc.Kind == "direct_ssot" && doc.HasLoop {
			count++
		}
	}
	return count
}

func directMissingLoops(docs []docClass) []string {
	var paths []string
	for _, doc := range docs {
		if doc.Kind == "direct_ssot" && !doc.HasLoop {
			paths = append(paths, doc.Path)
		}
	}
	if paths == nil {
		return []string{}
	}
	return paths
}
