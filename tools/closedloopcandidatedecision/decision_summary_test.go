package main

import (
	"strings"
	"testing"
)

func TestCandidateDecisionEvidenceExposesDecisionSummary(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	evidenceOut := t.TempDir() + "/evidence.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	m := manifestWithGeneratedCandidateRecord(t)
	manifestPath := t.TempDir() + "/manifest.json"
	if err := writeJSON(manifestPath, m); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: "../..", Manifest: manifestPath, CandidateIn: out, EvidenceOut: evidenceOut}); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got evidence
	if err := readJSON(evidenceOut, &got); err != nil {
		t.Fatal(err)
	}
	if got.DecisionSummary.RegisteredDecisionCount == 0 ||
		got.DecisionSummary.ConsumedDecisionCount != 1 {
		t.Fatalf("decision summary = %+v", got.DecisionSummary)
	}
	if summaryCountFor(got.DecisionSummary.DecisionSourceCounts, decisionSourceRecord) != 1 {
		t.Fatalf("decision source counts = %+v", got.DecisionSummary.DecisionSourceCounts)
	}
	if summaryCountFor(got.DecisionSummary.NextArtifactCounts, "claim_binding") == 0 {
		t.Fatalf("next artifact counts = %+v", got.DecisionSummary.NextArtifactCounts)
	}
}

func TestCandidateDecisionDocRendersDecisionSummary(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	doc := renderDoc(m, verifyResult{DecisionCount: len(m.Decisions)})
	for _, want := range []string{
		"## Decision Summary",
		"registered decisions",
		"next artifact counts",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("doc missing %q:\n%s", want, doc)
		}
	}
}
