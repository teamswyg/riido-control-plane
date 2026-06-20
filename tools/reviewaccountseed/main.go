package main

import (
	"flag"
	"os"
)

func main() {
	mustRun(os.Args[1:])
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("reviewaccountseed", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	manifestPath := fs.String("manifest", defaultManifest, "review seed manifest")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "verify generated reader doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(options{
		Repo: *repo, Manifest: *manifestPath, EvidenceOut: *evidenceOut,
		WriteDoc: *writeDoc, CheckDoc: *checkDoc,
	})
}
