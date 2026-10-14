package main

import (
	"flag"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/cloudwatchemf/requirements"
)

func main() {
	mustRun(os.Args[1:])
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("cloudwatchemf", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	manifest := fs.String("manifest", requirements.DefaultManifest, "CloudWatch EMF manifest")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "verify generated reader doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(options{
		Repo:        *repo,
		Manifest:    *manifest,
		EvidenceOut: *evidenceOut,
		WriteDoc:    *writeDoc,
		CheckDoc:    *checkDoc,
	})
}
