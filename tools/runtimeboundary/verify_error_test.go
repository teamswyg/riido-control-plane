package main

import (
	"strings"
	"testing"
)

func TestVerifyManifestShapeRejectsMissingFields(t *testing.T) {
	t.Parallel()
	err := verifyManifestShape(manifest{SchemaVersion: manifestSchema})
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("verifyManifestShape missing fields err = %v", err)
	}
}

func TestVerifyManifestShapeRejectsBadSchema(t *testing.T) {
	t.Parallel()
	err := verifyManifestShape(manifest{SchemaVersion: "wrong"})
	if err == nil || !strings.Contains(err.Error(), manifestSchema) {
		t.Fatalf("verifyManifestShape bad schema err = %v", err)
	}
}

func TestVerifyLoopRejectsIncompleteLoop(t *testing.T) {
	t.Parallel()
	err := verifyLoop(loopRecord{Observation: "observed", Hypothesis: "hypothesis", Execute: "run"})
	if err == nil || !strings.Contains(err.Error(), "evaluate") {
		t.Fatalf("verifyLoop incomplete err = %v", err)
	}
}

func TestVerifyEvidenceCheckRejectsMissingPhrase(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	writeFile(t, repo, "evidence.txt", "hello runtime")
	result := verifyResult{}
	err := verifyEvidenceCheck(repo, "boundary", evidenceCheck{
		Path:     "evidence.txt",
		Contains: []string{"missing"},
	}, &result)
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("verifyEvidenceCheck err = %v", err)
	}
	if result.EvidencePaths != 1 || result.PhraseChecks != 1 {
		t.Fatalf("result counters = %+v", result)
	}
}
