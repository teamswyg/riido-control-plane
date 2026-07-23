package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsLoopClosureAuditWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "loop-closure-audit.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("loop-closure-audit workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[13].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("loop-closure-audit digest substitution must fail closed")
	}
}

func TestVerifyRejectsLoopClosureAuditAdapterEnvironmentArtifactOrRetirementAuthority(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[13].NativeMapping.LoopClosureAudit.AdapterRequired = true
	document.BoundedChildren[13].ParityClaim.RequiredAdapterCount = 1
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("loop-closure-audit parity must remain native with zero adapters")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[13].ParityClaim.LoopClosureAuditLoopEnvironmentExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("loop-closure-audit loop environment must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[13].NativeMapping.EvidenceArtifact.Paths[0] = "out/substituted.json"
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("loop-closure-audit artifact order and paths must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[13].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[13].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[13].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("loop-closure-audit retirement requires separate owner review")
	}
}
