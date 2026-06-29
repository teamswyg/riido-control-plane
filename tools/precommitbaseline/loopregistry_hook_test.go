package main

import (
	"strings"
	"testing"
)

func TestPreCommitBaselineConfigRunsLoopRegistryHook(t *testing.T) {
	root := repoRootForTest(t)
	text, err := readText(repoPath(root, ".pre-commit-config.yaml"))
	if err != nil {
		t.Fatalf("read pre-commit config: %v", err)
	}
	phrases := []string{
		"repo: local",
		"id: loop-registry-claim-binding",
		"entry: tools/loopregistry/precommit.sh",
		"pass_filenames: false",
	}
	if err := requirePhrases(text, phrases, "loop-registry pre-commit hook", &verifyResult{}); err != nil {
		t.Fatal(err)
	}
}

func TestPreCommitBaselineLoopRegistryScriptRequiresImpactEvidence(t *testing.T) {
	root := repoRootForTest(t)
	text, err := readText(repoPath(root, "tools/loopregistry/precommit.sh"))
	if err != nil {
		t.Fatalf("read loopregistry precommit script: %v", err)
	}
	phrases := []string{
		"RIIDO_LOOP_REGISTRY_IMPACT_BASE",
		"RIIDO_LOOP_REGISTRY_EVIDENCE_OUT",
		"go run ./tools/loopregistry",
		"-check-doc",
		"-impact-base",
		"-target-verifier-summary",
		"-evidence-out",
	}
	if err := requirePhrases(text, phrases, "loopregistry pre-commit script", &verifyResult{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, " -write-hashes") {
		t.Fatal("pre-commit must not rewrite semantic hashes")
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}
