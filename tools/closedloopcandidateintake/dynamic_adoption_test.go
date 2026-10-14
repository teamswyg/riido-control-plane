package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCandidateIntakeAllowsSubjectBoundNextArtifact(t *testing.T) {
	root := repoRootForTest(t)
	out := candidateFixtureForTest(t, root)
	addDynamicArtifact(t, out, "custom_artifact", "go test ./custom", "custom_artifact")
	if _, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), out); err != nil {
		t.Fatal(err)
	}
}

func TestCandidateIntakeRejectsUnboundDynamicNextArtifact(t *testing.T) {
	root := repoRootForTest(t)
	out := candidateFixtureForTest(t, root)
	addDynamicArtifact(t, out, "custom_artifact", "go test ./custom", "other_artifact")
	_, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), out)
	if err == nil || !strings.Contains(err.Error(), "unknown next artifact custom_artifact") {
		t.Fatalf("expected unbound dynamic artifact failure, got %v", err)
	}
}

func addDynamicArtifact(t *testing.T, path, artifact, command, subjectArtifact string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var candidates candidateEvidence
	if err := json.Unmarshal(data, &candidates); err != nil {
		t.Fatal(err)
	}
	item := &candidates.Candidates[0]
	item.Subject = json.RawMessage(`{"kind":"test_subject","next_artifact":"` + subjectArtifact + `"}`)
	item.RequiredNextArtifacts = append([]string{artifact}, item.RequiredNextArtifacts...)
	item.AdoptionPlan = append([]adoptionStep{{Artifact: artifact, Command: command}}, item.AdoptionPlan...)
	data, err = json.MarshalIndent(candidates, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
