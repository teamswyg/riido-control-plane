package main

import (
	"strings"
	"testing"
)

func TestLoopRegistryEvidenceCarriesRefreshPlanSummary(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	if got.RefreshPlanSummary.PlanCount != len(got.RefreshPlans) {
		t.Fatalf("refresh summary = %+v, plans = %d", got.RefreshPlanSummary, len(got.RefreshPlans))
	}
	if got.RefreshPlanSummary.EvidenceArtifactCount == 0 ||
		got.RefreshPlanSummary.NextCommandCount == 0 ||
		got.RefreshPlanSummary.ManualCommandCount != len(got.RefreshPlans) {
		t.Fatalf("refresh summary missing counts: %+v", got.RefreshPlanSummary)
	}
}

func TestLoopRegistryDocRendersRefreshPlanSummary(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	doc := renderDoc(m, result)
	for _, want := range []string{
		"## Refresh Plan Summary",
		"evidence artifacts",
		"next commands",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("doc missing %q", want)
		}
	}
}
