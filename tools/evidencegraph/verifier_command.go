package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func verifierCommands(refs []ref) []string {
	seen := map[string]bool{}
	for _, item := range refs {
		pkg := verifierPackage(item)
		if pkg == "" {
			continue
		}
		seen[fmt.Sprintf("go test %s -count=1", pkg)] = true
	}
	out := make([]string, 0, len(seen))
	for command := range seen {
		out = append(out, command)
	}
	sort.Strings(out)
	return out
}

func verifierPackage(item ref) string {
	switch item.Kind {
	case "test", "tool", "code":
	default:
		return ""
	}
	path := strings.TrimSpace(item.Path)
	if path == "" || strings.HasPrefix(path, ".github/") {
		return ""
	}
	if strings.HasSuffix(path, ".go") {
		path = filepath.ToSlash(filepath.Dir(path))
	}
	if path == "." {
		return "."
	}
	return "./" + strings.TrimPrefix(path, "./")
}
