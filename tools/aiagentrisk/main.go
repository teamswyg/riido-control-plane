package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "aiagentrisk:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("aiagentrisk", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	manifestPath := fs.String("manifest", defaultManifest, "AI Agent risk evidence manifest path")
	repo := fs.String("repo", ".", "repository root")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	repoRoot, err := findRepoRoot(*repo)
	if err != nil {
		return err
	}
	loadPath := filepath.Join(repoRoot, filepath.FromSlash(*manifestPath))
	manifest, err := loadManifest(loadPath)
	if err != nil {
		return err
	}
	result, err := verifyManifest(repoRoot, manifest)
	if err != nil {
		return err
	}
	if *evidenceOut != "" {
		if err := writeEvidence(*evidenceOut, newEvidence(manifest, result)); err != nil {
			return err
		}
	}
	fmt.Printf("aiagentrisk: verified %d local evidence, %d external evidence, %d remaining boundaries\n", result.LocalEvidence, result.ExternalEvidence, result.RemainingBoundary)
	return nil
}
