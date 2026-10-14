package main

import (
	"go/ast"
	"strconv"
)

func collectConstants(file *ast.File, constants map[string]string) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok.String() != "const" {
			continue
		}
		for _, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if ok && len(valueSpec.Values) == len(valueSpec.Names) {
				collectValueSpec(valueSpec, constants)
			}
		}
	}
}

func collectValueSpec(spec *ast.ValueSpec, constants map[string]string) {
	for i, name := range spec.Names {
		lit, ok := spec.Values[i].(*ast.BasicLit)
		if !ok {
			continue
		}
		value, err := strconv.Unquote(lit.Value)
		if err == nil {
			constants[name.Name] = value
		}
	}
}

func isOSGetenv(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Getenv" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "os"
}

func collectEnvArg(expr ast.Expr, constants map[string]string, names map[string]bool) {
	switch arg := expr.(type) {
	case *ast.BasicLit:
		if value, err := strconv.Unquote(arg.Value); err == nil && envNamePattern.MatchString(value) {
			names[value] = true
		}
	case *ast.Ident:
		if value, ok := constants[arg.Name]; ok && envNamePattern.MatchString(value) {
			names[value] = true
		}
	}
}
