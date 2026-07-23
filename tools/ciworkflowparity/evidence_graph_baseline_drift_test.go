package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsEvidenceGraphWorkflowOrChildDigestDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, ".github", "workflows", "evidence-graph.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n# unreviewed drift\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("evidence-graph workflow drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.BoundedChildren[14].Baseline.WorkflowSHA256 = strings.Repeat("0", 64)
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("evidence-graph digest substitution must fail closed")
	}
}

func TestVerifyRejectsEvidenceGraphHistoryCacheImpactAnnotationArtifactOrAuthorityDrift(t *testing.T) {
	document := readRepositoryManifest(t)
	document.BoundedChildren[14].NativeMapping.Checkout.FullHistory = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("evidence-graph checkout must retain full history")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[14].NativeMapping.GoToolchain.Cache = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("evidence-graph Go cache behavior must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[14].ParityClaim.EvidenceGraphImpactBaseExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("evidence-graph optional impact base must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[14].ParityClaim.EvidenceGraphAnnotationsExact = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("evidence-graph annotations must remain exact")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[14].NativeMapping.EvidenceArtifact.RunWhen = "always"
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("evidence-graph artifact must remain success-only")
	}
	document = readRepositoryManifest(t)
	document.BoundedChildren[14].Authority.WorkflowRetirementAuthorized = true
	document.BoundedChildren[14].Authority.WorkflowFileEffect = "delete"
	document.BoundedChildren[14].Rollback.BaselineWorkflowPreserved = false
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("evidence-graph retirement requires separate owner review")
	}
}
