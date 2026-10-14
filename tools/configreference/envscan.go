package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

var envNamePattern = regexp.MustCompile(`^(RIIDO|AWS|OTEL)_[A-Z0-9_]+$`)

func scanRuntimeEnv(repoRoot, sourceDir string) (map[string]bool, error) {
	names := map[string]bool{}
	constants := map[string]string{}
	files, err := filepath.Glob(repoPath(repoRoot, filepath.ToSlash(filepath.Join(sourceDir, "*.go"))))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		if err := scanGoFile(file, constants, names); err != nil {
			return nil, err
		}
	}
	for _, value := range constants {
		if envNamePattern.MatchString(value) {
			names[value] = true
		}
	}
	return names, nil
}

func scanGoFile(file string, constants map[string]string, names map[string]bool) error {
	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		return err
	}
	collectConstants(parsed, constants)
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if ok && isOSGetenv(call) && len(call.Args) == 1 {
			collectEnvArg(call.Args[0], constants, names)
		}
		return true
	})
	return nil
}
