package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/requirements"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "saascontrolplane:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("saascontrolplane", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	manifest := fs.String("manifest", requirements.DefaultManifest, "SaaS control-plane manifest")
	boundary := fs.String("boundary", "", "optional boundary evidence id")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "check generated reader doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(options{Repo: *repo, Manifest: *manifest, Boundary: *boundary, EvidenceOut: *evidenceOut, WriteDoc: *writeDoc, CheckDoc: *checkDoc})
}
