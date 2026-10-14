package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsConfigReferenceWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "config-reference.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("config-reference workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[7].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("config-reference digest substitution must fail closed")
	}
}

func TestVerifyRejectsConfigReferenceAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[7].NativeMapping.ConfigReference.AdapterRequired = true
	document.BoundedChildren[7].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("config-reference parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[7].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[7].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[7].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("config-reference retirement requires separate owner review")
	}
}
