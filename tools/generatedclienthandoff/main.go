package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "generatedclienthandoff:", err)
		os.Exit(1)
	}
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.OpenAPI, "openapi", "", "OpenAPI JSON path")
	flag.StringVar(&cfg.DSL, "dsl", "", "DSL JSON path")
	flag.StringVar(&cfg.IR, "ir", "", "IR JSON path")
	flag.StringVar(&cfg.Core, "core", "", "generated core TypeScript path")
	flag.StringVar(&cfg.React, "react", "", "generated React TypeScript path")
	flag.StringVar(&cfg.Out, "out", "", "output directory")
	flag.StringVar(&cfg.PRBody, "pr-body", "", "optional generated PR body path")
	flag.StringVar(&cfg.PreviousManifest, "previous-manifest", "", "optional previous manifest path")
	flag.StringVar(&cfg.SourceRepo, "source-repo", "teamswyg/riido-control-plane", "source repository")
	flag.StringVar(&cfg.SourceRef, "source-ref", "", "source ref or tag")
	flag.StringVar(&cfg.SourceCommit, "source-commit", "", "source commit SHA")
	flag.StringVar(&cfg.TargetRepo, "target-repo", "teamswyg/riido-client", "target repository")
	flag.StringVar(&cfg.TargetBranch, "target-branch", "", "target branch name")
	flag.StringVar(&cfg.GeneratedAt, "generated-at", "", "YYYY-MM-DD generated date")
	flag.Parse()
	return cfg
}
