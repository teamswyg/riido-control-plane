package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "evidencegraph:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("evidencegraph", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	opt := options{}
	fs.StringVar(&opt.Repo, "repo", ".", "repository root")
	fs.StringVar(&opt.Manifest, "manifest", defaultManifest, "evidence graph manifest")
	fs.StringVar(&opt.EvidenceOut, "evidence-out", "", "optional evidence JSON output")
	fs.StringVar(&opt.ImpactBase, "impact-base", "", "optional git base ref for PR impact verification")
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated reader doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "verify generated reader doc")
	verify := fs.Bool("verify", false, "alias for -check-doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opt.CheckDoc = opt.CheckDoc || *verify
	return run(opt)
}
