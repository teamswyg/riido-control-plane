package main

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

func scanForbiddenImports(repoRoot string, packages []packageEntry, forbidden []string) ([]string, error) {
	var violations []string
	for _, pkg := range packages {
		files, err := filepath.Glob(repoPath(repoRoot, pkg.Path+"/*.go"))
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			next, err := forbiddenImportsInFile(file, forbidden)
			if err != nil {
				return nil, err
			}
			violations = append(violations, next...)
		}
	}
	return violations, nil
}

func forbiddenImportsInFile(path string, forbidden []string) ([]string, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var violations []string
	for _, item := range parsed.Imports {
		value, err := strconv.Unquote(item.Path.Value)
		if err == nil && matchesForbiddenImport(value, forbidden) {
			violations = append(violations, path+": "+value)
		}
	}
	return violations, nil
}

func matchesForbiddenImport(value string, forbidden []string) bool {
	for _, fragment := range forbidden {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}
