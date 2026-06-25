package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCandidateDecisionEvidenceNamesConsumedArtifact(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	result, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err != nil {
		t.Fatal(err)
	}
	ev := newEvidence(loadDecisionManifestForTest(t), result)
	if len(ev.ConsumedArtifacts) != 1 {
		t.Fatalf("expected one consumed artifact, got %d", len(ev.ConsumedArtifacts))
	}
	artifact := ev.ConsumedArtifacts[0]
	if artifact.InputPath != out {
		t.Fatalf("input path mismatch: %q", artifact.InputPath)
	}
	if artifact.SourceGeneratedAt == "" || artifact.SourceExpiresAt == "" {
		t.Fatalf("consumed artifact freshness fields are required: %+v", artifact)
	}
	if len(artifact.CandidateIDs) == 0 {
		t.Fatalf("consumed artifact must bind candidate ids: %+v", artifact)
	}
}

func TestCandidateDecisionCLIEvidenceNamesConsumedArtifact(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	candidateIn := dir + "/candidates.json"
	evidenceOut := dir + "/evidence.json"
	if err := generateCandidate(t, root, candidateIn); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CandidateIn: candidateIn, EvidenceOut: evidenceOut}); err != nil {
		t.Fatal(err)
	}
	ev := readEvidenceForTest[evidence](t, evidenceOut)
	if len(ev.ConsumedArtifacts) != 1 {
		t.Fatalf("expected CLI evidence to expose consumed artifact: %+v", ev)
	}
}

func readEvidenceForTest[T any](t *testing.T, path string) T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
