package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func requireFile(root, path string) error {
	_, err := os.Stat(repoPath(root, path))
	return err
}

func testSymbols(root string) (map[string]bool, error) {
	symbols := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		return collectTests(path, symbols)
	})
	return symbols, err
}

func collectTests(path string, symbols map[string]bool) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "Test") {
			symbols[fn.Name.Name] = true
		}
	}
	return nil
}
