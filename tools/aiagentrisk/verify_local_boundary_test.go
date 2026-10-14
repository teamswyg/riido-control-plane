package main

import (
	"strings"
	"testing"
)

func TestVerifyLocalEvidenceRejectsInvalidShapeAndRisk(t *testing.T) {
	err := verifyLocalEvidence(t.TempDir(), localEvidence{}, "")
	if err == nil || !strings.Contains(err.Error(), "invalid local evidence") {
		t.Fatalf("expected invalid local evidence, got %v", err)
	}

	item := localEvidence{
		Risk: "unknown", Status: "verified", Package: "./tools/aiagentrisk",
		Test: "TestDefaultManifestVerifies", Proves: "risk evidence",
	}
	err = verifyLocalEvidence("../..", item, item.Test)
	if err == nil || !strings.Contains(err.Error(), "invalid local evidence risk/test") {
		t.Fatalf("expected invalid risk/test, got %v", err)
	}
}

func TestVerifyBoundariesRejectsInvalidUnexpectedAndMissing(t *testing.T) {
	tests := []struct {
		name       string
		boundaries []remainingBoundary
		want       string
	}{
		{"invalid", []remainingBoundary{{ID: requiredBoundaries[0]}}, "invalid remaining boundary"},
		{
			"unexpected",
			[]remainingBoundary{
				{ID: requiredBoundaries[0], Owner: "codex", Reason: "known"},
				{ID: "other", Owner: "codex", Reason: "unknown"},
			},
			"unexpected remaining boundary",
		},
		{
			"missing",
			[]remainingBoundary{{ID: requiredBoundaries[0], Owner: "codex", Reason: "known"}},
			"manifest missing remaining boundary",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyBoundaries(tc.boundaries)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}
}
