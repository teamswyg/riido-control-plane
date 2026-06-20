package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("liveworkflowevidence", flag.ContinueOnError)
	opt := options{}
	fs.StringVar(&opt.Repo, "repo", ".", "repository root")
	fs.StringVar(&opt.Manifest, "manifest", defaultManifest, "manifest path")
	fs.StringVar(&opt.WorkflowID, "workflow", "", "live workflow id")
	fs.StringVar(&opt.LiveStatus, "live-status", "", "redacted live status")
	fs.StringVar(&opt.DeploymentTarget, "deployment-target", "", "deployment target")
	fs.StringVar(&opt.EvidenceOut, "evidence-out", "", "optional evidence JSON output path")
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "verify generated doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(opt)
}
