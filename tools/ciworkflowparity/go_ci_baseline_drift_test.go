package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsGoCIWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "go-ci.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("Go CI workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[2].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("Go CI digest substitution must fail closed")
	}
}

func TestVerifyRejectsGoCIAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[2].NativeMapping.Lint.AdapterRequired = true
	document.BoundedChildren[2].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("Go CI parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[2].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[2].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[2].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("Go CI retirement requires separate owner review")
	}
}
