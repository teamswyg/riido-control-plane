package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsContextMapWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "context-map.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("context map workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[1].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("context map digest substitution must fail closed")
	}
}

func TestVerifyRejectsContextMapAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[1].NativeMapping.ContextMap.AdapterRequired = true
	document.BoundedChildren[1].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("context map parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[1].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[1].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[1].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("context map retirement requires separate owner review")
	}
}
