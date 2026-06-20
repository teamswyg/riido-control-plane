package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "repositoryreadme:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("repositoryreadme", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "repository root")
	manifest := fs.String("manifest", defaultManifest, "README manifest")
	writeDoc := fs.Bool("write-doc", false, "write generated README")
	checkDoc := fs.Bool("check-doc", false, "verify generated README")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return runWithOptions(*repo, *manifest, *writeDoc, *checkDoc, *evidenceOut)
}
