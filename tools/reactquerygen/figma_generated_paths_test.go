package main

import (
	"sort"
	"strings"
)

func canonicalPathFromFigmaFacade(path string) string {
	out := strings.TrimPrefix(path, "riido.")
	out = strings.TrimPrefix(out, "v2.")
	return out
}

func generatedPathsByOperation(spec openAPISpec) map[string]string {
	out := map[string]string{}
	for path, methods := range spec.Paths {
		for method, operation := range methods {
			generatedPath := operation.Client.GeneratedPath
			if strings.TrimSpace(generatedPath) == "" {
				generatedPath = generatedPathFromClient(operation.Client)
			}
			out[generatedPath] = strings.ToUpper(method) + " " + path + " " + operation.OperationID
		}
	}
	return out
}

func generatedPathHaystack(spec openAPISpec, generatedPaths map[string]string) string {
	parts := make([]string, 0, len(generatedPaths)+len(spec.Paths))
	for generatedPath, route := range generatedPaths {
		parts = append(parts, generatedPath+" "+route)
	}
	sort.Strings(parts)
	return strings.ToLower(strings.Join(parts, "\n"))
}

func docMentionsGeneratedPath(docText, generatedPath string) bool {
	if strings.Contains(docText, generatedPath) {
		return true
	}
	lastDot := strings.LastIndex(generatedPath, ".")
	if lastDot < 0 {
		return false
	}
	return strings.Contains(docText, generatedPath[:lastDot]+".*")
}
