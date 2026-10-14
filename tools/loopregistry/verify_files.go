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

func testSymbols(root string) (map[string][]string, error) {
	symbols := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		pkg, err := packagePathForFile(root, path)
		if err != nil {
			return err
		}
		return collectTests(path, pkg, symbols)
	})
	return symbols, err
}

func collectTests(path, pkg string, symbols map[string][]string) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "Test") {
			symbols[fn.Name.Name] = append(symbols[fn.Name.Name], pkg)
		}
	}
	return nil
}
