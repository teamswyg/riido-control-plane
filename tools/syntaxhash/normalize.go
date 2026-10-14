package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

func normalizedFile(path string) (string, string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return "", "", err
	}
	decls := make([]string, 0, len(file.Decls))
	for _, decl := range file.Decls {
		decls = append(decls, declShape(decl))
	}
	sort.Strings(decls)
	shape := strings.Join(decls, "|")
	return filepath.Base(path) + ":" + shape, shape, nil
}

func fieldCount(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}
	total := 0
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			total++
			continue
		}
		total += len(field.Names)
	}
	return total
}
