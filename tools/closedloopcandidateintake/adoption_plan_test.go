package main

import (
	"strings"
	"testing"
)

func TestCandidateIntakeRejectsMissingAdoptionPlan(t *testing.T) {
	m := loadIntakeManifestForTest(t)
	artifact := intakeTestCandidateEvidence()
	item := closedLoopCandidate{
		ID: "smoke:broken", HarnessLoop: "provider_acceptance_harness",
		SourceRef:       intakeTestSourceRef(m),
		PromotionTarget: m.Sources[0].PromotionTarget, Observation: "broken",
		PromotionEdge: graphEdge{
			From: "provider_acceptance_harness", To: m.Sources[0].PromotionTarget, Relation: "promotes_failure_to",
		},
		RequiredNextArtifacts: m.Sources[0].RequiredNextArtifacts,
	}
	_, err := verifyCandidateItem(m, artifact, item)
	if err == nil || !strings.Contains(err.Error(), "adoption plan") {
		t.Fatalf("expected adoption plan failure, got %v", err)
	}
}

func intakeTestCandidateEvidence() candidateEvidence {
	return candidateEvidence{
		SourceWorkflow:    ".github/workflows/smoke.yml",
		LiveStatus:        "failure",
		SourceGeneratedAt: "2026-06-24T00:00:00Z",
		SourceExpiresAt:   "2026-06-25T00:00:00Z",
		CandidateCount:    1,
	}
}

func intakeTestSourceRef(m manifest) candidateSourceRef {
	return candidateSourceRef{
		HarnessLoop:       m.Sources[0].HarnessLoop,
		SourceWorkflow:    ".github/workflows/smoke.yml",
		SummaryArtifact:   "smoke-summary",
		CandidateArtifact: m.Sources[0].CandidateArtifact,
		LiveStatus:        "failure",
		SourceGeneratedAt: "2026-06-24T00:00:00Z",
		SourceExpiresAt:   "2026-06-25T00:00:00Z",
		Run:               runRecord{ID: "run-1"},
	}
}
