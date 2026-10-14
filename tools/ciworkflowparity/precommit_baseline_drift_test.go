package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsPreCommitWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "pre-commit-baseline.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("pre-commit-baseline workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[4].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("pre-commit-baseline digest substitution must fail closed")
	}
}

func TestVerifyRejectsPreCommitAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[4].NativeMapping.PreCommitBaseline.AdapterRequired = true
	document.BoundedChildren[4].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("pre-commit-baseline parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[4].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[4].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[4].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("pre-commit-baseline retirement requires separate owner review")
	}
}
