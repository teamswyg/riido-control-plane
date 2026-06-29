package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "operationalreadiness:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("operationalreadiness", flag.ContinueOnError)
	opt := options{}
	fs.StringVar(&opt.Repo, "repo", ".", "repository root")
	fs.StringVar(&opt.Manifest, "manifest", defaultManifest, "readiness manifest")
	fs.StringVar(&opt.EvidenceOut, "evidence-out", "", "optional evidence JSON output")
	fs.StringVar(&opt.CandidateOut, "candidate-out", "", "optional closed-loop candidate JSON output")
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "verify generated doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(opt)
}
