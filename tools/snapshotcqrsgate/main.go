package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "snapshotcqrsgate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("snapshotcqrsgate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	manifestPath := fs.String("manifest", defaultManifest, "snapshot CQRS gate manifest")
	repo := fs.String("repo", ".", "repository root")
	checkDoc := fs.Bool("check-doc", false, "verify the paired reader doc")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	repoRoot, err := findRepoRoot(*repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(filepath.Join(repoRoot, filepath.FromSlash(*manifestPath)))
	if err != nil {
		return err
	}
	result, err := verify(repoRoot, m, *checkDoc)
	if err != nil {
		return err
	}
	if *evidenceOut != "" {
		if err := writeEvidence(*evidenceOut, newEvidence(m, result)); err != nil {
			return err
		}
	}
	fmt.Printf("snapshotcqrsgate: verified %d operations, %d signals, %d rules\n", result.Operations, result.Signals, result.DecisionRules)
	return nil
}
