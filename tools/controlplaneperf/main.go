package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "controlplaneperf:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("controlplaneperf", flag.ContinueOnError)
	opt := options{}
	fs.StringVar(&opt.Repo, "repo", ".", "repository root")
	fs.StringVar(&opt.Manifest, "manifest", defaultManifest, "performance manifest")
	fs.StringVar(&opt.EvidenceOut, "evidence-out", "", "optional evidence JSON output")
	fs.StringVar(&opt.ArchitectureQueryOut, "architecture-query-out", "", "optional architecture query JSON output")
	fs.Func("architecture-query-path", "architecture file path to route", func(value string) error {
		opt.ArchitecturePaths = append(opt.ArchitecturePaths, value)
		return nil
	})
	fs.BoolVar(&opt.WriteDoc, "write-doc", false, "write generated doc")
	fs.BoolVar(&opt.CheckDoc, "check-doc", false, "verify generated doc")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(opt)
}
