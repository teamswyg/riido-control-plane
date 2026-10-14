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

func TestCandidateIntakeEvidenceExposesAdoptionPlan(t *testing.T) {
	root := repoRootForTest(t)
	path := candidateFixtureForTest(t, root)
	result, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CandidateAdoptionPlans) != result.CandidateCount {
		t.Fatalf("expected adoption plan evidence for each candidate, got %d/%d",
			len(result.CandidateAdoptionPlans), result.CandidateCount)
	}
	plan := result.CandidateAdoptionPlans[0]
	if plan.CandidateID == "" || len(plan.RequiredNextArtifacts) == 0 || len(plan.AdoptionPlan) == 0 {
		t.Fatalf("adoption plan evidence is incomplete: %+v", plan)
	}
	if plan.AdoptionPlan[0].Artifact == "" || plan.AdoptionPlan[0].Command == "" {
		t.Fatalf("adoption plan evidence must carry executable command: %+v", plan.AdoptionPlan[0])
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
