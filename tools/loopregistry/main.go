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
	verify := fs.Bool("verify", false, "alias for -check-doc")
	writeHashes := fs.Bool("write-hashes", false, "update claim semantic hashes")
	impactBase := fs.String("impact-base", "", "optional git base ref for PR impact verification")
	githubAnnotations := fs.Bool("github-annotations", false, "emit GitHub Actions claim verifier annotations")
	targetSummary := fs.Bool("target-verifier-summary", false, "emit local target verifier plan summary")
	refreshPlanIn := fs.String("refresh-plan-in", "", "existing loop registry evidence JSON")
	refreshCommandsOut := fs.String("refresh-commands-out", "", "selected refresh command JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(options{
		Repo:               *repo,
		Manifest:           *manifest,
		EvidenceOut:        *evidenceOut,
		WriteDoc:           *writeDoc,
		CheckDoc:           *checkDoc || *verify,
		WriteHashes:        *writeHashes,
		ImpactBase:         *impactBase,
		GitHubAnnotations:  *githubAnnotations,
		TargetSummary:      *targetSummary,
		RefreshPlanIn:      *refreshPlanIn,
		RefreshCommandsOut: *refreshCommandsOut,
	})
}
