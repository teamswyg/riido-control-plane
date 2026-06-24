package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "loopregistry:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("loopregistry", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "repository root")
	manifest := fs.String("manifest", defaultManifest, "loop registry manifest")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "verify generated reader doc")
	writeHashes := fs.Bool("write-hashes", false, "update claim semantic hashes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(options{*repo, *manifest, *evidenceOut, *writeDoc, *checkDoc, *writeHashes})
}
