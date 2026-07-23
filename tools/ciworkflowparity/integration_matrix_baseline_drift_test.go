package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsIntegrationMatrixWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "integration-matrix.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("integration-matrix workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[12].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("integration-matrix digest substitution must fail closed")
	}
}

func TestVerifyRejectsIntegrationMatrixAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[12].NativeMapping.IntegrationMatrix.AdapterRequired = true
	document.BoundedChildren[12].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("integration-matrix parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[12].ParityClaim.ExecutableKnowledgeCommandExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("integration-matrix executable-knowledge behavior must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[12].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[12].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[12].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("integration-matrix retirement requires separate owner review")
	}
}
