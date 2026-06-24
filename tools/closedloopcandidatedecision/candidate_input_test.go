package main

import (
	"errors"
	"testing"
)

func TestCandidateDecisionVerifyRequiresCandidateInput(t *testing.T) {
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
