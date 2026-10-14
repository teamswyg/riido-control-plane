package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyManifestShapeRejectsMissingRequiredFields(t *testing.T) {
	if err := verifyManifestShape(manifest{}); err == nil {
		t.Fatalf("expected required field error")
	}
	m := baseManifest()
	m.Loop.Execute = ""
	if err := verifyManifestShape(m); err == nil || !strings.Contains(err.Error(), "loop") {
		t.Fatalf("expected loop error, got %v", err)
	}
}

func TestVerifyPublicGateRejectsInvalidDefinitions(t *testing.T) {
	root := tempRepo(t)
	result := verifyResult{}
	err := verifyPublicGate(root, nil, publicGate{}, &result)
	if err == nil || !strings.Contains(err.Error(), "public gate") {
		t.Fatalf("expected public gate field error, got %v", err)
	}
	gate := publicGate{
		Surface:            "pull",
		Verification:       "uses aws manually",
		ExternalDependency: "aws",
		PullRequestGate:    true,
	}
	if err := verifyPublicGate(root, []string{"aws"}, gate, &result); err == nil {
		t.Fatalf("expected forbidden dependency error")
	}
	gate.Verification = "no aws dependency"
	gate.Workflows = []string{"missing.yml"}
	if err := verifyPublicGate(root, []string{"aws"}, gate, &result); err == nil {
		t.Fatalf("expected missing workflow error")
	}
}

func TestVerifyPublicGatesRejectsDuplicatesAndMissingPullRequest(t *testing.T) {
	root := tempRepo(t)
	m := baseManifest()
	m.PublicGates = append(m.PublicGates, m.PublicGates[0])
	if err := verifyPublicGates(root, m, &verifyResult{}); err == nil {
		t.Fatalf("expected duplicate public gate error")
	}
	workflow := filepath.Join(root, ".github", "workflows", "matrix.yml")
	if err := os.MkdirAll(filepath.Dir(workflow), 0o755); err != nil {
		t.Fatalf("mkdir workflow: %v", err)
	}
	if err := os.WriteFile(workflow, []byte("on: push\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	gate := publicGate{
		Surface: "pull", Verification: "check", ExternalDependency: "none",
		PullRequestGate: true, Workflows: []string{".github/workflows/matrix.yml"},
	}
	if err := verifyPublicGate(root, nil, gate, &verifyResult{}); err == nil {
		t.Fatalf("expected missing pull_request trigger error")
	}
}
