package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "closedloopcandidatedecision:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("closedloopcandidatedecision", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	opt := options{}
	fs.StringVar(&opt.Repo, "repo", ".", "repository root")
	fs.StringVar(&opt.Manifest, "manifest", defaultManifest, "decision manifest")
	fs.StringVar(&opt.CandidateIn, "candidate-in", "", "closed-loop candidate input")
	fs.StringVar(&opt.EvidenceOut, "evidence-out", "", "optional evidence JSON output")
	fs.StringVar(&opt.CommandsOut, "commands-out", "", "optional loop refresh command evidence output")
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "verify generated doc")
	verify := fs.Bool("verify", false, "alias for -check-doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opt.CheckDoc = opt.CheckDoc || *verify
	return run(opt)
}
