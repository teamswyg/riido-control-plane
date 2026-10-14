package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsRepositoryReadmeWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "repository-readme.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("repository README workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[0].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("repository README digest substitution must fail closed")
	}
}

func TestVerifyRejectsRepositoryReadmeAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[0].NativeMapping.ExecutableKnowledge.AdapterRequired = true
	document.BoundedChildren[0].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("repository README parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[0].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[0].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[0].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("repository README retirement requires separate owner review")
	}
}
