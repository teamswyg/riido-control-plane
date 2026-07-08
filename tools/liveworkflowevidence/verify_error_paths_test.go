package main

import (
	"strings"
	"testing"
)

func TestVerifyWorkflowSpecErrorPaths(t *testing.T) {
	t.Parallel()
	m := liveTestManifest()
	if err := verifyWorkflowSpec(workflowSpec{}, map[string]bool{}); err == nil {
		t.Fatal("expected identity error")
	}
	spec := m.Workflows[0]
	if err := verifyWorkflowSpec(spec, map[string]bool{spec.ID: true}); err == nil {
		t.Fatal("expected duplicate workflow id error")
	}
	spec.SensitiveInputs = nil
	if err := verifyWorkflowSpec(spec, map[string]bool{}); err == nil {
		t.Fatal("expected sensitive input error")
	}
	spec = m.Workflows[0]
	spec.EvidenceTTLHours = 0
	if err := verifyWorkflowSpec(spec, map[string]bool{}); err == nil {
		t.Fatal("expected ttl error")
	}
}

func TestVerifyClaimSpecErrorPaths(t *testing.T) {
	t.Parallel()
	spec := liveTestManifest().Workflows[0]
	spec.AllowedFields = []string{"generated_at", "expires_at"}
	if err := verifyWorkflowSpec(spec, map[string]bool{}); err == nil ||
		!strings.Contains(err.Error(), "evidence_claims") {
		t.Fatalf("expected evidence_claims field error, got %v", err)
	}
	spec = liveTestManifest().Workflows[0]
	spec.EvidenceClaims[0].SourcePhrases = nil
	if err := verifyClaimSpecs(spec); err == nil {
		t.Fatal("expected source phrase error")
	}
	spec = liveTestManifest().Workflows[0]
	spec.EvidenceClaims = append(spec.EvidenceClaims, spec.EvidenceClaims[0])
	if err := verifyClaimSpecs(spec); err == nil {
		t.Fatal("expected duplicate claim error")
	}
}
