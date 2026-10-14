package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCandidateIntakePreservesCandidateSubject(t *testing.T) {
	root := repoRootForTest(t)
	out := candidateFixtureForTest(t, root)
	writeFirstCandidateSubject(t, out, `{"kind":"claim_coverage_gap","claim_id":"claim_one"}`)
	result, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), out)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CandidateSubjects) != 1 ||
		result.CandidateSubjects[0].Kind != "claim_coverage_gap" ||
		!strings.Contains(string(result.CandidateSubjects[0].Subject), "claim_one") {
		t.Fatalf("candidate subjects = %+v", result.CandidateSubjects)
	}
	ev := newEvidence(loadIntakeManifestForTest(t), result)
	if len(ev.CandidateSubjects) != 1 {
		t.Fatalf("evidence subjects = %+v", ev.CandidateSubjects)
	}
}

func TestCandidateIntakeRejectsSubjectWithoutKind(t *testing.T) {
	root := repoRootForTest(t)
	out := candidateFixtureForTest(t, root)
	writeFirstCandidateSubject(t, out, `{"claim_id":"claim_one"}`)
	_, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), out)
	if err == nil || !strings.Contains(err.Error(), "subject must include kind") {
		t.Fatalf("expected subject kind failure, got %v", err)
	}
}

func writeFirstCandidateSubject(t *testing.T, path, subject string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var artifact candidateEvidence
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatal(err)
	}
	artifact.Candidates[0].Subject = json.RawMessage(subject)
	data, err = json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
