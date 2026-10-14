package main

import (
	"encoding/json"
	"testing"
)

func TestHarnessFailureCandidateCarriesSubject(t *testing.T) {
	source := promotionSource{
		ID: "smoke", HarnessLoop: "provider_acceptance_harness",
		SourceWorkflow: ".github/workflows/smoke.yml", SummaryArtifact: "summary",
		CandidateArtifact: "candidates", PromotionTarget: "closed_loop_candidate",
		FailureStatuses: []string{"failure"},
	}
	summary := liveSummary{
		ID: "smoke", LiveStatus: "failure",
		EvidenceClaims: []liveClaim{{ID: "provider_smoke", Status: "not_verified"}},
	}
	got := buildCandidateEvidence(source, summary).Candidates[0]
	var subject candidateSubject
	if err := json.Unmarshal(got.Subject, &subject); err != nil {
		t.Fatal(err)
	}
	if subject.Kind != "harness_failure" ||
		subject.ClaimID != "provider_smoke" ||
		subject.HarnessLoop != source.HarnessLoop ||
		subject.LiveStatus != "failure" {
		t.Fatalf("subject = %+v", subject)
	}
}
