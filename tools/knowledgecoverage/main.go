package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	manifest := flag.String("manifest", "docs/executable-knowledge.riido.json", "coverage manifest")
	evidenceOut := flag.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := flag.Bool("write-doc", false, "write generated reader doc")
	checkDoc := flag.Bool("check-doc", false, "check generated reader doc")
	flag.Parse()
	if err := run(*repo, *manifest, *evidenceOut, *writeDoc, *checkDoc); err != nil {
		fmt.Fprintln(os.Stderr, "knowledgecoverage:", err)
		os.Exit(1)
	}
	fmt.Println("knowledgecoverage: verified")
}
