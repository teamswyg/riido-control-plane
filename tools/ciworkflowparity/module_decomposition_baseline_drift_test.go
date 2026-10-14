package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsModuleDecompositionWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "module-decomposition.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("module-decomposition workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[3].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("module-decomposition digest substitution must fail closed")
	}
}

func TestVerifyRejectsModuleDecompositionAdapterOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[3].NativeMapping.ModuleDecomposition.AdapterRequired = true
	document.BoundedChildren[3].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("module-decomposition parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[3].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[3].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[3].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("module-decomposition retirement requires separate owner review")
	}
}
