package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "runtimecdownership:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("runtimecdownership", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "repository root")
	manifestPath := fs.String("manifest", defaultManifest, "runtime CD ownership manifest")
	writeDoc := fs.Bool("write-doc", false, "write the generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "verify the generated reader doc")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return runWithOptions(*repo, *manifestPath, *writeDoc, *checkDoc, *evidenceOut)
}
