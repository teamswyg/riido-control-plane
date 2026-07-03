package main

import (
	"fmt"
	"go/ast"
	"strings"
)

func declShape(decl ast.Decl) string {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return funcShape(d)
	case *ast.GenDecl:
		parts := make([]string, 0, len(d.Specs))
		for _, spec := range d.Specs {
			parts = append(parts, specShape(spec))
		}
		return fmt.Sprintf("gen:%s:%s", d.Tok.String(), strings.Join(parts, ","))
	default:
		return fmt.Sprintf("decl:%T", decl)
	}
}

func funcShape(fn *ast.FuncDecl) string {
	body := "nil"
	if fn.Body != nil {
		parts := []string{}
		for _, stmt := range fn.Body.List {
			parts = append(parts, fmt.Sprintf("%T", stmt))
		}
		body = strings.Join(parts, ",")
	}
	return fmt.Sprintf("func:recv=%d:name=%s:params=%d:results=%d:body=%s",
		fieldCount(fn.Recv), fn.Name.Name, fieldCount(fn.Type.Params),
		fieldCount(fn.Type.Results), body)
}

func specShape(spec ast.Spec) string {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		return fmt.Sprintf("type:%s:%T", s.Name.Name, s.Type)
	case *ast.ValueSpec:
		return fmt.Sprintf("value:names=%d:values=%d:type=%T",
			len(s.Names), len(s.Values), s.Type)
	case *ast.ImportSpec:
		return "import"
	default:
		return fmt.Sprintf("spec:%T", spec)
	}
}
