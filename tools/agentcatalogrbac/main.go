package main

import (
	"flag"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/agentcatalogrbac/requirements"
)

func main() {
	mustRun(os.Args[1:])
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("agentcatalogrbac", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	manifest := fs.String("manifest", requirements.DefaultManifest, "agent catalog RBAC manifest")
	profile := fs.String("profile", "rbac", "evidence profile id")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "check generated reader doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(options{Repo: *repo, Manifest: *manifest, Profile: *profile, EvidenceOut: *evidenceOut, WriteDoc: *writeDoc, CheckDoc: *checkDoc})
}
