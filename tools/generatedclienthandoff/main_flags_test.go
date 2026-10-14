package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlagsMapsCLIValues(t *testing.T) {
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()
	flag.CommandLine = flag.NewFlagSet("generatedclienthandoff-test", flag.ContinueOnError)
	os.Args = []string{
		"generatedclienthandoff",
		"-openapi", "openapi.json",
		"-dsl", "contract.dsl.json",
		"-ir", "contract.ir.json",
		"-core", "core.ts",
		"-react", "react.ts",
		"-out", "out",
		"-pr-body", "PR.md",
		"-previous-manifest", "previous.ts",
		"-source-repo", "source/repo",
		"-source-ref", "v1",
		"-source-commit", "abc",
		"-target-repo", "target/repo",
		"-target-branch", "RIID-1",
		"-generated-at", "2026-07-09",
	}
	got := parseFlags()
	if got.OpenAPI != "openapi.json" || got.TargetRepo != "target/repo" {
		t.Fatalf("parsed config = %+v", got)
	}
	if got.PreviousManifest != "previous.ts" || got.GeneratedAt != "2026-07-09" {
		t.Fatalf("parsed optional config = %+v", got)
	}
}
