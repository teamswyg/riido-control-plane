package main

import (
	"flag"
	"fmt"
	"os"
)

const defaultManifest = "docs/30-architecture/syntax-hash.riido.json"

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "syntaxhash:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("syntaxhash", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	opt := options{}
	fs.StringVar(&opt.Repo, "repo", ".", "repository root")
	fs.StringVar(&opt.Manifest, "manifest", defaultManifest, "syntax hash manifest")
	fs.StringVar(&opt.EvidenceOut, "evidence-out", "", "optional evidence output")
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "check generated doc")
	verify := fs.Bool("verify", false, "alias for -check-doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opt.CheckDoc = opt.CheckDoc || *verify
	return run(opt)
}

func verifyGraph(m manifest, graph syntaxGraph) error {
	if graph.Repository.CoverageBasisPoints < m.Constraints.MinRepositoryCoverageBasisPoint {
		return fmt.Errorf("repository coverage %d bp below %d bp",
			graph.Repository.CoverageBasisPoints, m.Constraints.MinRepositoryCoverageBasisPoint)
	}
	for i, target := range graph.Targets {
		if err := verifyTarget(m, graph, i, target); err != nil {
			return err
		}
	}
	return nil
}

func verifyTarget(m manifest, graph syntaxGraph, i int, target targetGraph) error {
	if m.Targets[i].GoldenCommand == "" {
		return fmt.Errorf("target %s missing golden command", target.ID)
	}
	if target.Coverage < m.Constraints.MinCoveragePercent {
		return fmt.Errorf("target %s coverage %d below %d",
			target.ID, target.Coverage, m.Constraints.MinCoveragePercent)
	}
	if target.PackageHash == "" || target.SemanticHash == "" {
		return fmt.Errorf("target %s missing syntax or semantic hash", target.ID)
	}
	if target.GoFiles != target.TrackedFiles {
		return fmt.Errorf("target %s has untracked go files", target.ID)
	}
	if len(target.Relocations) != len(target.FileHashes) {
		return fmt.Errorf("target %s missing relocation mapping", target.ID)
	}
	return nil
}
