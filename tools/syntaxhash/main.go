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
