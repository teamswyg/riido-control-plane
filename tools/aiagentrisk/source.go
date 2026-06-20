package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

func testFunctionExists(repoRoot, packagePath, testName string) (bool, error) {
	if !strings.HasPrefix(packagePath, "./") || strings.Contains(packagePath, "..") {
		return false, errUnsafePackagePath(packagePath)
	}
	root := filepath.Join(repoRoot, strings.TrimPrefix(packagePath, "./"))
	found := false
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		ok, err := fileContainsTestFunction(path, testName)
		if err != nil {
			return err
		}
		found = found || ok
		return nil
	})
	return found, err
}

func fileContainsTestFunction(path, testName string) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return false, err
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == testName {
			return true, nil
		}
	}
	return false, nil
}
