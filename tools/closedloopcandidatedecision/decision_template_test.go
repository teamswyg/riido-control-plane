package main

import (
	"strings"
	"testing"
)

func TestCandidateDecisionUsesTemplateForIgnoredCommandSubject(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/ignored-candidates.json"
	command := "go test ./tools/looprefreshdispatch"
	if err := writeJSON(out, ignoredCommandCandidateArtifact(t, command)); err != nil {
		t.Fatal(err)
	}
	pinCandidateFreshnessClock(t)
	result, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	if len(result.DecisionArtifacts) != 1 {
		t.Fatalf("decision artifacts = %+v", result.DecisionArtifacts)
	}
	artifact := result.DecisionArtifacts[0]
	if artifact.NextArtifact != "verifier" || artifact.NextCommand != command {
		t.Fatalf("template decision artifact = %+v", artifact)
	}
}

func TestCandidateDecisionRejectsTemplateWithoutSubjectKind(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	m.DecisionTemplates[0].SubjectKind = ""
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing template subject kind to fail")
	}
}

func TestCandidateDecisionRejectsUnknownTemplateNextLoop(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	m.DecisionTemplates[0].NextLoop = "missing_loop"
	_, err := verifyAll("../..", m)
	if err == nil || !strings.Contains(err.Error(), "decision template") {
		t.Fatalf("expected unknown template next loop failure, got %v", err)
	}
}
