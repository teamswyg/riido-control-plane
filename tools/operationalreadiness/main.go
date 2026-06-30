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
	fs.StringVar(&opt.PublicStatusOut, "public-status-out", "", "optional public status Markdown output")
	fs.StringVar(&opt.PublicStatusHTML, "public-status-html-out", "", "optional public status HTML output")
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "verify generated doc")
	verify := fs.Bool("verify", false, "alias for -check-doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opt.CheckDoc = opt.CheckDoc || *verify
	return run(opt)
}
