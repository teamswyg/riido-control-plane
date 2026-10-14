package main

import (
	"flag"
	"os"
)

func main() {
	mustRun(os.Args[1:])
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("contextmap", flag.ContinueOnError)
	return runWithFlags(fs, args)
}

func runWithFlags(fs *flag.FlagSet, args []string) error {
	opts := options{}
	fs.StringVar(&opts.Repo, "repo", ".", "repository root")
	fs.StringVar(&opts.Manifest, "manifest", defaultManifest, "context map manifest path")
	fs.StringVar(&opts.EvidenceOut, "evidence-out", "", "optional evidence JSON output path")
	fs.BoolVar(&opts.WriteDoc, "write-doc", false, "write generated reader doc")
	fs.BoolVar(&opts.CheckDoc, "check-doc", false, "check generated reader doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(opts)
}
