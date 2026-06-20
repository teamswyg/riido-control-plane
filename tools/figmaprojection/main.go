package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "figmaprojection:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("figmaprojection", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "repository root")
	projection := fs.String("projection", defaultProjection, "projection manifest")
	source := fs.String("source", defaultSource, "mirrored source manifest")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "run the generated-client projection gate")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(*repo, *projection, *source, *evidenceOut, *writeDoc, *checkDoc)
}
