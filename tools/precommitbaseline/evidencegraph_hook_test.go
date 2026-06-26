package main

import (
	"strings"
	"testing"
)

func TestPreCommitBaselineConfigRunsEvidenceGraphHook(t *testing.T) {
	root := repoRootForTest(t)
	text, err := readText(repoPath(root, ".pre-commit-config.yaml"))
	if err != nil {
		t.Fatalf("read pre-commit config: %v", err)
	}
	phrases := []string{
		"repo: local",
		"id: evidence-graph-chain-binding",
		"entry: tools/evidencegraph/precommit.sh",
		"pass_filenames: false",
	}
	if err := requirePhrases(text, phrases, "evidencegraph pre-commit hook", &verifyResult{}); err != nil {
		t.Fatal(err)
	}
}

func TestPreCommitBaselineEvidenceGraphScriptRequiresImpactEvidence(t *testing.T) {
	root := repoRootForTest(t)
	text, err := readText(repoPath(root, "tools/evidencegraph/precommit.sh"))
	if err != nil {
		t.Fatalf("read evidencegraph precommit script: %v", err)
	}
	phrases := []string{
		"RIIDO_EVIDENCE_GRAPH_IMPACT_BASE",
		"RIIDO_EVIDENCE_GRAPH_EVIDENCE_OUT",
		"go run ./tools/evidencegraph",
		"-check-doc",
		"-impact-base",
		"-evidence-out",
	}
	if err := requirePhrases(text, phrases, "evidencegraph pre-commit script", &verifyResult{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, " -write-doc") {
		t.Fatal("pre-commit must not rewrite evidence graph docs")
	}
}
