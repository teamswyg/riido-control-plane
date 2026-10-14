package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsSyntaxHashWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "syntax-hash.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("syntax-hash workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[6].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("syntax-hash digest substitution must fail closed")
	}
}

func TestVerifyRejectsSyntaxHashEnvironmentAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[6].ParityClaim.SyntaxHashLoopEnvironmentExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("syntax-hash loop environment drift must fail closed")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[6].NativeMapping.SyntaxHash.AdapterRequired = true
	document.BoundedChildren[6].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("syntax-hash parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[6].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[6].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[6].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("syntax-hash retirement requires separate owner review")
	}
}
