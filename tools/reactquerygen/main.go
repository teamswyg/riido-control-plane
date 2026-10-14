package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	openAPIPath := flag.String("openapi", "", "OpenAPI JSON path")
	outPath := flag.String("out", "", "generated TypeScript path")
	flag.Parse()
	if err := run(*openAPIPath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, "reactquerygen:", err)
		os.Exit(1)
	}
}
