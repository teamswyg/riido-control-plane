package main

import (
	"strings"
	"testing"
)

func TestVerifyAllRejectsMissingIdentityAndWorkflows(t *testing.T) {
	t.Parallel()
	if _, err := verifyAll(t.TempDir(), manifest{}); err == nil ||
		!strings.Contains(err.Error(), "identity") {
		t.Fatalf("expected identity error, got %v", err)
	}
	m := liveTestManifest()
	m.Workflows = nil
	if _, err := verifyAll(t.TempDir(), m); err == nil ||
		!strings.Contains(err.Error(), "live workflow") {
		t.Fatalf("expected workflow error, got %v", err)
	}
}

func TestVerifyClaimSpecsRejectsIncompleteClaim(t *testing.T) {
	t.Parallel()
	spec := liveTestManifest().Workflows[0]
	spec.EvidenceClaims[0].Summary = ""
	if err := verifyClaimSpecs(spec); err == nil ||
		!strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("expected incomplete claim error, got %v", err)
	}
}
