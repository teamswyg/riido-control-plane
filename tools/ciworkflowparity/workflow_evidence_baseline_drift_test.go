package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsWorkflowEvidenceWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "workflow-evidence.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("workflow-evidence workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[9].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("workflow-evidence digest substitution must fail closed")
	}
}

func TestVerifyRejectsWorkflowEvidenceAdapterEnvironmentOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[9].NativeMapping.WorkflowEvidence.AdapterRequired = true
	document.BoundedChildren[9].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("workflow-evidence parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[9].ParityClaim.WorkflowEvidenceLoopEnvironmentExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("workflow-evidence loop environment must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[9].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[9].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[9].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("workflow-evidence retirement requires separate owner review")
	}
}
