package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	openAPIPath := flag.String("openapi", "", "OpenAPI JSON path")
	outPath := flag.String("out", "", "generated TypeScript path")
	typesOnly := flag.Bool("types-only", false, "emit only OpenAPI schema types")
	dslPath := flag.String("dsl", "", "DSL JSON path to regenerate IR and OpenAPI")
	flag.Parse()
	if *dslPath != "" {
		generateContract(strings.TrimSuffix(*dslPath, ".dsl.riido.json"))
		return
	}
	if *typesOnly {
		spec, err := loadOpenAPI(*openAPIPath)
		if err != nil {
			panic(err)
		}
		var b strings.Builder
		writeGeneratedHeader(&b)
		writeTypes(&b, spec.Components.Schemas)
		if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
			panic(err)
		}
		if err := os.WriteFile(*outPath, []byte(b.String()), 0o644); err != nil {
			panic(err)
		}
		return
	}
	if err := run(*openAPIPath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, "reactquerygen:", err)
		os.Exit(1)
	}
}
