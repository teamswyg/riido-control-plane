package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsLegacyWorkflowOrBaselineDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "ci.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("legacy workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("baseline digest substitution must fail closed")
	}
}

func TestVerifyRejectsAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.NativeMapping.EvidenceArtifact.AdapterRequired = true
	document.ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("baseline parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.Authority.WorkflowRetirementAuthorized = true
	document.Authority.WorkflowFileEffect = "delete"
	document.Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("workflow retirement requires a separate owner-reviewed change")
	}
}
