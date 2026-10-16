package main

import (
	"os"
	"strings"
	"testing"
)

func assertPendingReviewCoverage(t *testing.T, m manifest, review pendingReview) {
	t.Helper()
	sources := map[string]bool{}
	for _, source := range review.Sources {
		if source.ID == "" || sources[source.ID] {
			t.Fatal("review source identity must be unique")
		}
		if source.Kind == "repository_file" {
			if _, err := os.Stat(repoPath(repoRootForTest(t), source.Path)); err != nil {
				t.Fatalf("review evidence path: %v", err)
			}
		}
		sources[source.ID] = true
	}
	items := map[string]reviewedDecision{}
	for _, item := range review.Items {
		if _, duplicate := items[item.CandidateID]; duplicate {
			t.Fatal("duplicate reviewed candidate")
		}
		if item.Observation == "" || item.RemainingGap == "" || item.NextStep == "" || len(item.EvidenceRefs) == 0 {
			t.Fatal("review needs actual observations, retained gap, evidence, and next step")
		}
		for _, ref := range item.EvidenceRefs {
			if !sources[ref] {
				t.Fatalf("unknown review evidence %q", ref)
			}
		}
		items[item.CandidateID] = item
	}
	var pending []decisionRecord
	for _, decision := range m.Decisions {
		if decisionNeedsReviewBy(decision.Disposition) {
			pending = append(pending, decision)
		}
	}
	for _, template := range m.DecisionTemplates {
		pending = append(pending, decisionFromTemplate("template:"+template.SubjectKind, template))
	}
	if len(items) != len(pending) {
		t.Fatal("every pending decision and template needs exactly one review")
	}
	for _, decision := range pending {
		assertReviewedDecision(t, decision, review.ReviewBy, items[decision.CandidateID])
	}
}

func assertReviewedDecision(t *testing.T, decision decisionRecord, due string, item reviewedDecision) {
	t.Helper()
	if decision.CandidateID != item.CandidateID || decision.Disposition != item.Disposition ||
		decision.Priority != item.Priority || decision.Owner != item.Owner || decision.NextLoop != item.NextLoop ||
		decision.NextArtifact != item.NextArtifact || decision.ReviewBy != due || item.ReviewBy != due {
		t.Fatalf("review decision mismatch for %s", decision.CandidateID)
	}
	if !strings.Contains(decision.Reason, pendingReviewRef+decision.CandidateID) {
		t.Fatalf("decision must cite its item-level review: %s", decision.CandidateID)
	}
}
