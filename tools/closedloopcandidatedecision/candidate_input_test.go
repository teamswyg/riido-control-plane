package main

import (
	"errors"
	"testing"
)

func TestCandidateDecisionVerifyRequiresCandidateInput(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-07-15T00:00:00Z")
	err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	})
	if !errors.Is(err, errMissingCandidateInput) {
		t.Fatalf("expected missing candidate input, got %v", err)
	}
}
