package main

import "testing"

func TestCandidateDecisionRejectsUnknownNextArtifact(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	m := loadDecisionManifestForTest(t)
	m.Decisions[0].NextArtifact = "spreadsheet"
	if _, err := verifyCandidateDecisions(root, m, out); err == nil {
		t.Fatal("expected unknown next_artifact to fail")
	}
}

func TestCandidateDecisionEvidenceBindsNextArtifactCommand(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	result, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	if len(result.DecisionArtifacts) != 1 {
		t.Fatalf("decision artifacts = %+v", result.DecisionArtifacts)
	}
	if result.DecisionArtifacts[0].NextArtifact != "claim_binding" {
		t.Fatalf("next artifact = %q", result.DecisionArtifacts[0].NextArtifact)
	}
	if result.DecisionArtifacts[0].NextCommand == "" {
		t.Fatalf("next command missing: %+v", result.DecisionArtifacts[0])
	}
}

func TestCandidateDecisionEvidenceFileIncludesNextArtifactCommand(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	evidenceOut := t.TempDir() + "/evidence.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CandidateIn: out, EvidenceOut: evidenceOut}); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got evidence
	if err := readJSON(evidenceOut, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.DecisionArtifacts) != 1 || got.DecisionArtifacts[0].NextArtifact != "claim_binding" {
		t.Fatalf("decision artifacts = %+v", got.DecisionArtifacts)
	}
	if got.DecisionArtifacts[0].NextCommand == "" {
		t.Fatalf("next command missing: %+v", got.DecisionArtifacts[0])
	}
	if got.DecisionArtifacts[0].PromotionEdge.Relation != "promotes_failure_to" {
		t.Fatalf("promotion edge missing: %+v", got.DecisionArtifacts[0])
	}
}
