package main

import (
	"flag"
	"os"
)

func main() {
	mustRun(os.Args[1:])
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("aiagentclientapi", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	manifest := fs.String("manifest", defaultManifest, "AI Agent client API manifest")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "check generated reader doc")
	verify := fs.Bool("verify", false, "alias for -check-doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opts := options{Repo: *repo, Manifest: *manifest, EvidenceOut: *evidenceOut}
	opts.WriteDoc = *writeDoc
	opts.CheckDoc = *checkDoc || *verify
	return run(opts)
}
