package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/requirements"
	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/runconfig"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "apiclientdelivery:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("apiclientdelivery", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "repository root")
	manifestPath := fs.String("manifest", requirements.DefaultManifest, "API client delivery manifest")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "verify generated reader doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(runconfig.Options{Repo: *repo, Manifest: *manifestPath, EvidenceOut: *evidenceOut, WriteDoc: *writeDoc, CheckDoc: *checkDoc})
}
