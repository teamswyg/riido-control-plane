package aiagentrisk

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func assertTestFunctionExists(t *testing.T, packagePath, testName string) {
	t.Helper()
	if !strings.HasPrefix(packagePath, "./") || strings.Contains(packagePath, "..") {
		t.Fatalf("unsafe package path %q", packagePath)
	}

	root := filepath.Clean(filepath.Join("../..", strings.TrimPrefix(packagePath, "./")))
	found := false
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		if fileContainsTestFunction(t, path, testName) {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk package %s: %v", packagePath, err)
	}
	if !found {
		t.Fatalf("%s does not contain %s", packagePath, testName)
	}
}

func fileContainsTestFunction(t *testing.T, path, testName string) bool {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == testName {
			return true
		}
	}
	return false
}
