package main

import "testing"

func TestCandidateDecisionEvidenceMarksRecordDecisionSource(t *testing.T) {
	result := generatedCandidateDecisionResult(t, manifestWithGeneratedCandidateRecord(t))
	if len(result.DecisionArtifacts) != 1 {
		t.Fatalf("decision artifacts = %+v", result.DecisionArtifacts)
	}
	artifact := result.DecisionArtifacts[0]
	if artifact.DecisionSource != decisionSourceRecord {
		t.Fatalf("decision source = %+v", artifact)
	}
	if artifact.DecisionTemplateSubjectKind != "" {
		t.Fatalf("template subject kind = %+v", artifact)
	}
}

func TestCandidateDecisionEvidenceMarksTemplateDecisionSource(t *testing.T) {
	result := ignoredCommandTemplateResult(t, "go test ./tools/looprefreshdispatch")
	if len(result.DecisionArtifacts) != 1 {
		t.Fatalf("decision artifacts = %+v", result.DecisionArtifacts)
	}
	artifact := result.DecisionArtifacts[0]
	if artifact.DecisionSource != decisionSourceTemplate ||
		artifact.DecisionTemplateSubjectKind != "loop_refresh_ignored_command" {
		t.Fatalf("template provenance = %+v", artifact)
	}
}

func generatedCandidateDecisionResult(t *testing.T, m manifest) verifyResult {
	t.Helper()
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	result, err := verifyCandidateDecisions(root, m, out)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	return result
}

func ignoredCommandTemplateResult(t *testing.T, command string) verifyResult {
	t.Helper()
	root := repoRootForTest(t)
	out := t.TempDir() + "/ignored-candidates.json"
	if err := writeJSON(out, ignoredCommandCandidateArtifact(t, command)); err != nil {
		t.Fatal(err)
	}
	pinCandidateFreshnessClock(t)
	result, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	return result
}
